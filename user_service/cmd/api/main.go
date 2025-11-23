// @title PocketPay API
// @version 1.0
// @description PocketPay API
// @host localhost:8000
// @BasePath /
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" // match Dockerfile EXPOSE
	}
	fmt.Println("PORT:", port)
	ctx := context.Background()
	pool, err := openDBPool(ctx, db_url)
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

	api.POST("/createUser", func(c *gin.Context) {
		var in createAccount

		err := c.ShouldBindJSON(&in)
		var id int
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}
		ctx := c.Request.Context()
		var exists bool

		temp_err := pool.QueryRow(
			ctx,
			"SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)",
			in.Email,
		).Scan(&exists)
		if temp_err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": temp_err.Error()})
		}
		if exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "The user already exists"})
		}
		error := pool.QueryRow(ctx, "INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id", in.Email, in.Password).Scan(&id)
		if error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": error.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": id})

	})

	api.POST("/updateUser", func(c *gin.Context) {
		var in updateUser

		err := c.ShouldBindJSON(&in)

		var id int

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}

		ctx := c.Request.Context()
		error := pool.QueryRow(ctx, "UPDATE  users set password = $2 where email = $1 RETURNING id", in.Email, in.Password).Scan(&id)
		if error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": error.Error()})
		}
		c.JSON(http.StatusOK, gin.H{"id": id})
	})

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		_ = srv.ListenAndServe()
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	fmt.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("server forced to shutdown: %v\n", err)
	}
	fmt.Println("server exited")
}
