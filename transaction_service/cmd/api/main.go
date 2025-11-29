package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PayAnotherAccountRequest struct {
	FromAccountEmail string `json:"from_account" binding:"required,email"`
	ToAccountEmail   string `json:"to_account" binding:"required,email"`
	Amount           int    `json:"amount" binding:"required,min=0.1"`
}

func openDBPool(ctx context.Context, db_url string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(db_url)
	if err != nil {
		return nil, fmt.Errorf("parse db url: %w", err)

	}
	config.MaxConns = 5
	config.MinConns = 1
	config.MaxConnLifetime = time.Minute * 30
	pool, error := pgxpool.NewWithConfig(ctx, config)
	if error != nil {
		return nil, fmt.Errorf("create db pool error: %w", err)
	}
	return pool, nil

}
func main() {
	db_url := os.Getenv("DB_url")
	r := gin.New()
	ctx := context.Background()
	pool, err := openDBPool(ctx, db_url)

	api := r.Group("/api/v1")
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	api.POST("/pay_another_account", func(c *gin.Context) {
		var in PayAnotherAccountRequest

		bind_err := c.ShouldBindBodyWithJSON(&in)

		if bind_err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"details": "Incorrect request format"})
			return
		}
		ctx := c.Request.Context()
		const transferSQL = `
			UPDATE account_balance
			SET balance = CASE
				WHEN email = $1 THEN balance - $3
				WHEN email = $2 THEN balance + $3
			END
			WHERE email IN ($1, $2)
			AND (email != $1 OR balance - $3 >= 0);
			`
		_, db_err := pool.Exec(ctx, transferSQL, in.FromAccountEmail, in.ToAccountEmail, in.Amount)

		if db_err != nil {
			fmt.Println(db_err)
			c.JSON(http.StatusInternalServerError, gin.H{"details": "Incorrect request format"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"details": "Amount was sent"})

	})

}
