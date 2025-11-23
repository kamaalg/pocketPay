// @title PocketPay API
// @version 1.0
// @description PocketPay API
// @host localhost:8001
// @BasePath /
package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type AddCardRequest struct {
	CardNo  string `json:"cardNo"   binding:"required,numeric,len=16"`
	ExpDate string `json:"exp_date" binding:"required"`
	Cvc     string `json:"cvc"      binding:"required,numeric,len=3"`
}

// i
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" // match Dockerfile EXPOSE
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
	})

	_ = r.Run(":" + port)
}
