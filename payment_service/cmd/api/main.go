// @title PocketPay API
// @version 1.0
// @description PocketPay API
// @host localhost:8001
// @BasePath /
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	dbpackage "github.com/kamaalg/pocketPay/db"
)

type AddCardRequest struct {
	CardNo  string `json:"cardNo"   binding:"required,numeric,len=16"`
	ExpDate string `json:"exp_date" binding:"required"`
	Cvc     string `json:"cvc"      binding:"required,numeric,len=3"`
}

type DeleteCardRequest struct {
	CardNo string `json:"CardNo" binding:"required,numeric,len=16"`
}

var pool *pgxpool.Pool

// // authMiddleware requires an Authorization header. If AUTH_TOKEN env var is set,
// // the header must be "Bearer <AUTH_TOKEN>". If AUTH_TOKEN is empty, the header
// // must simply be present and use the Bearer scheme.
// func authMiddleware() gin.HandlerFunc {
// 	expected := os.Getenv("AUTH_TOKEN")
// 	return func(c *gin.Context) {
// 		auth := c.GetHeader("Authorization")
// 		if auth == "" {
// 			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
// 			return
// 		}
// 		// expect "Bearer <token>"
// 		parts := strings.Fields(auth)
// 		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
// 			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header must be Bearer token"})
// 			return
// 		}
// 		token := parts[1]
// 		if expected != "" && token != expected {
// 			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
// 			return
// 		}
// 		// token ok (or presence-only mode)
// 		c.Next()
// 	}
// }

// i
func main() {
	port := os.Getenv("PORT")
	db_url := os.Getenv("DB_URL")
	ctx := context.Background()
	pool, _ = dbpackage.OpenDBPool(ctx, db_url)
	if port == "" {
		port = "8001" // match Dockerfile EXPOSE
	}

	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"welcome": "hello"})
	})
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "service healthy"})
	})

	api := r.Group("api/v1")

	api.POST("/add_card", func(c *gin.Context) {
		var in AddCardRequest
		var exists bool
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"details": "Wrong request format", "error": err.Error()})
			return
		}
		query_err := pool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM card_details WHERE cardNO = $1)", in.CardNo).Scan(&exists)

		if query_err != nil {
			fmt.Println(query_err)
			c.JSON(http.StatusInternalServerError, gin.H{"details": "DB error."})
			return
		}

		if exists {
			c.JSON(http.StatusOK, gin.H{"details": "The card already exists in our records."})
		}

		_, insert_err := pool.Exec(ctx, "INSERT INTO card_details (cardNO, exp_date, cvv) VALUES ($1,$2,$3)", in.CardNo, in.ExpDate, in.Cvc)

		if insert_err != nil {
			fmt.Println(query_err)
			c.JSON(http.StatusInternalServerError, gin.H{"details": "DB error."})
			return
		}
		// TODO: process card (tokenize, store, forward to provider)
		c.JSON(http.StatusOK, gin.H{"status": "card added"})
	})

	api.DELETE("/delete_card", func(c *gin.Context) {
		var in DeleteCardRequest
		var exists bool
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"details": "Wrong request format", "error": err.Error()})
			return
		}
		query_err := pool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM card_details WHERE cardNO = $1)", in.CardNo).Scan(&exists)

		if query_err != nil {
			fmt.Println(query_err)
			c.JSON(http.StatusInternalServerError, gin.H{"details": "DB error."})
			return
		}

		if exists == false {
			c.JSON(http.StatusOK, gin.H{"details": "The card does not exists in our records."})
		}

		_, delete_err := pool.Exec(ctx, "DELETE FROM card_details WHERE cardno=$1", in.CardNo)

		if delete_err != nil {
			fmt.Println(query_err)
			c.JSON(http.StatusInternalServerError, gin.H{"details": "DB error."})
			return
		}
		// TODO: process card (tokenize, store, forward to provider)
		c.JSON(http.StatusOK, gin.H{"status": "card deleted"})
	})

	_ = r.Run(":" + port)
}
