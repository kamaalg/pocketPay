package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=0"`
}
type SignUpRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=0"`
	Name     string `json:"name" binding:"required,min=0"`
	Age      int    `json:"age"`
}
type accessTokenClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func openDBpool(ctx context.Context, dburl string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dburl)
	if err != nil {
		return nil, fmt.Errorf("parse db url: %w", err)

	}
	config.MaxConns = 5
	config.MinConns = 1
	config.MaxConnLifetime = time.Minute * 30
	pool, err := pgxpool.NewWithConfig(ctx, config)
	return pool, nil
}
func generatetoken(email string) (accessTokenString string, refreshTokenString string, error error) {
	secret_token := os.Getenv("Secret_Token")
	jwtSecret := []byte(secret_token)
	accessTokenExpiry := time.Now().Add(45 * time.Minute)
	claims := accessTokenClaims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessTokenExpiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   email,
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessTokenString, err := accessToken.SignedString(jwtSecret)
	if err != nil {
		return "", "", err
	}
	refreshTokenBytes := make([]byte, 32)
	if _, err := rand.Read(refreshTokenBytes); err != nil {
		return "", "", err
	}
	refreshTokenStrin := hex.EncodeToString(refreshTokenBytes)
	return accessTokenString, refreshTokenStrin, nil

}
func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8002"
	}
	db_url := os.Getenv("DB_URL")
	ctx := context.Background()
	pool, err := openDBpool(ctx, db_url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to db: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()
	fmt.Println("auth service starting on url", db_url)

	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	api := r.Group("api/v1")

	api.POST("/login", func(c *gin.Context) {
		var in loginRequest
		err := c.ShouldBindBodyWithJSON(&in)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		}
		inserted_time := time.Now()
		fmt.Println("Current Time:", inserted_time)
		accessToken, refreshToken, err_token := generatetoken(in.Email)
		if err_token != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "problem with generating token"})
		}
		fmt.Println(accessToken, refreshToken)
		ctx := c.Request.Context()
		var storedHash string
		err_db_query := pool.QueryRow(ctx, "SELECT password FROM user_info WHERE email = $1", in.Email).Scan(&storedHash)
		if err_db_query != nil {
			if errors.Is(err_db_query, pgx.ErrNoRows) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid email or password",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "Error with db"})
			return
		}

		fmt.Println(storedHash)

		if storedHash == "" {
			c.JSON(http.StatusOK, "Email not found")
			return
		}
		bcrypt_err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(in.Password))

		if bcrypt_err != nil {
			c.JSON(http.StatusForbidden, gin.H{"details": "Password is not correct"})
			return
		}

		_, err_db := pool.Exec(ctx, "INSERT INTO auth_details (email, access_token,refresh_token,inserted_time) VALUES ($1, $2,$3,$4)", in.Email, accessToken, refreshToken, inserted_time)
		if err_db != nil {
			fmt.Println("Some error with db")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err_db.Error()})
			return

		}
		responseData := gin.H{
			"message":      "Success!!!",
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
		}
		c.JSON(http.StatusOK, responseData)

	})

	api.POST("/signup", func(c *gin.Context) {
		var in SignUpRequest
		json_err := c.ShouldBindBodyWithJSON(&in)

		if json_err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect request format"})
			return
		}

		created_time := time.Now()
		fmt.Println("Current Time:", created_time)
		ctx := c.Request.Context()
		var exists bool
		err_db := pool.QueryRow(ctx, "SELECT EXISTS  (SELECT 1 FROM user_info WHERE email = $1)", in.Email).Scan(&exists)
		if err_db != nil {
			fmt.Println(err_db.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB is not working"})
			return

		}

		if exists {
			c.JSON(http.StatusOK, gin.H{"details": "User with this email already exists, please sign in"})
			return
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to process password"})
			return
		}

		_, err_db_2 := pool.Exec(ctx, "INSERT INTO user_info (email,age,name,password) VALUES ($1,$2,$3,$4)", in.Email, in.Age, in.Name, hashedPassword)

		if err_db_2 != nil {
			fmt.Println(err_db_2.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB is not working"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success"})

	})

	// simple run (blocking)
	_ = r.Run(":" + port)
}
