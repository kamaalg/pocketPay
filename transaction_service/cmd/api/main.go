package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	dbpackage "github.com/kamaalg/pocketPay/db"
)

type PayAnotherAccountRequest struct {
	FromAccountEmail string `json:"from_account" binding:"required,email"`
	ToAccountEmail   string `json:"to_account" binding:"required,email"`
	Amount           int64  `json:"amount" binding:"required,min=0.1"`
	Idempotency_key  string `json:"idempotency_key" binding:"required"`
}

func main() {
	db_url := os.Getenv("DB_url")
	r := gin.New()
	ctx := context.Background()
	pool, err := dbpackage.OpenDBPool(ctx, db_url)

	if err != nil {
		fmt.Println(err)
		return
	}

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
