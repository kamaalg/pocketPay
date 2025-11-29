// @title PocketPay API
// @version 1.0
// @description PocketPay API
// @host localhost:8000
// @BasePath /
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	// import generated swagger docs (created by `swag init`)
)

type createAccount struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=0"`
}
type updateUser struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=0"`
}

type ChangeBalanceRequest struct {
	Amount int    `json:"amount" binding:"required,min=0.1"`
	Email  string `json:"email" binding:"required,email"`
}

func openDBPool(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dbURL)

	if err != nil {
		return nil, fmt.Errorf("parse db url: %w", err)

	}
	config.MaxConns = 10
	config.MinConns = 1
	config.MaxConnLifetime = time.Minute * 30

	pool, err := pgxpool.NewWithConfig(ctx, config)
	return pool, nil
}

func main() {
	db_url := os.Getenv("DB_URL")
	card_db_url := os.Getenv("Card_DB_URL")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" // match Dockerfile EXPOSE
	}
	fmt.Println("PORT:", port)
	ctx := context.Background()
	pool, err := openDBPool(ctx, db_url)
	card_pool, err := openDBPool(ctx, card_db_url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to db: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"welcome": "hello"})
	})
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "service healthy"})
	})
	api := r.Group("/api/v1")

	api.GET("/get_balance", func(c *gin.Context) {
		var balance int
		email := c.Query("email") // returns "" if not present

		if email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"details": "email query param is required"})
			return
		}
		ctx := c.Request.Context()

		db_err := pool.QueryRow(ctx, "SELECT balance FROM account_balance WHERE email=$1", email).Scan(&balance)

		if db_err != nil {
			fmt.Println(db_err)
			c.JSON(http.StatusInternalServerError, gin.H{"details": "DB error."})
			return
		}
		c.JSON(http.StatusOK, gin.H{"balance": balance})

	})

	api.GET("/get_user_info", func(c *gin.Context) {
		var name string
		var age int
		var password string
		email := c.Query("email") // returns "" if not present

		if email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"details": "email query param is required"})
			return
		}
		ctx := c.Request.Context()

		db_err := pool.QueryRow(ctx, "SELECT name,age,password FROM user_info WHERE email=$1", email).Scan(&name, &age, &password)

		if db_err != nil {
			fmt.Println(db_err)
			c.JSON(http.StatusInternalServerError, gin.H{"details": "DB error."})
			return
		}
		response_data := gin.H{
			"Age":      age,
			"Email":    email,
			"Name":     name,
			"Password": password,
		}
		c.JSON(http.StatusOK, response_data)

	})

	api.POST("/add_balance", func(c *gin.Context) {
		//EMULATE STRIPE API CALL
		var in ChangeBalanceRequest
		var exists bool

		ctx := c.Request.Context()

		bind_err := c.ShouldBindBodyWithJSON(&in)
		if bind_err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"details": "Wrong request format"})
			return
		}
		db_err := pool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM user_info WHERE email=$1)", in.Email).Scan(&exists)
		if db_err != nil {
			fmt.Println(db_err)
			c.JSON(http.StatusInternalServerError, gin.H{"details": "DB error."})
			return
		}
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"details": "No email matched with the provided one"})
			return
		}
		var newAmount int
		err_insert := card_pool.QueryRow(ctx,
			"UPDATE account_balance SET balance = balance + $1 WHERE email = $2 RETURNING balance",
			in.Amount,
			in.Email,
		).Scan(&newAmount)

		if err_insert != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"details": "DB error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"details": "Success,amount added", "New Amount": newAmount})

	})
	api.POST("/remove_balance", func(c *gin.Context) {
		var in ChangeBalanceRequest
		var exists bool
		ctx := c.Request.Context()

		bind_err := c.ShouldBindBodyWithJSON(&in)
		if bind_err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"details": "Wrong request format"})
			return
		}
		db_err := pool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM user_info WHERE email=$1)", in.Email).Scan(&exists)
		if db_err != nil {
			fmt.Println(db_err)
			c.JSON(http.StatusInternalServerError, gin.H{"details": "DB error."})
			return
		}
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"details": "No email matched with the provided one"})
			return
		}
		var newAmount int
		err_insert := card_pool.QueryRow(ctx,
			"UPDATE account_balance SET balance = balance - $1 WHERE email = $2 AND balance >= $1 RETURNING balance",
			in.Amount,
			in.Email,
		).Scan(&newAmount)

		if err_insert != nil {
			if errors.Is(err, pgx.ErrNoRows) {

				c.JSON(http.StatusBadRequest, gin.H{"details": "Insufficient balance"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"details": "DB error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"details": "Success,amount added", "New Amount": newAmount})
	})

	_ = r.Run(":" + port)
}
