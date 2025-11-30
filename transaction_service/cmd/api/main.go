package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	dbpackage "github.com/kamaalg/pocketPay/db"
	ledgerpb "github.com/kamaalg/pocketPay/ledger_service/ledgerpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	// Create a gRPC connection to ledger service once at startup
	ledgerAddr := os.Getenv("LEDGER_ADDR") // e.g. "ledger:50051" or "localhost:50051"
	var ledgerClient ledgerpb.LedgerClient
	if ledgerAddr != "" {
		conn, err := grpc.Dial(ledgerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Printf("failed to dial ledger service: %v\n", err)
		} else {
			ledgerClient = ledgerpb.NewLedgerClient(conn)
			// Note: we don't close conn here because we want the client to live for process lifetime.
			// If you prefer, store conn and close it on shutdown.
		}
	}

	api.POST("/pay_another_account", func(c *gin.Context) {
		var in PayAnotherAccountRequest

		bind_err := c.ShouldBindJSON(&in)

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

		// Optionally record the transfer in the ledger service via gRPC.
		if ledgerClient != nil {
			// Use idempotency key if provided; otherwise generate a simple id.
			txID := in.Idempotency_key
			if txID == "" {
				txID = fmt.Sprintf("tx-%d", time.Now().UnixNano())
			}

			ledgerReq := &ledgerpb.Transaction{
				Id:           txID,
				AccountEmail: in.FromAccountEmail,
				Amount:       in.Amount,
				Description:  fmt.Sprintf("transfer to %s", in.ToAccountEmail),
				Timestamp:    time.Now().Unix(),
			}

			// small timeout for the RPC to avoid blocking the request for long
			rpcCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			ack, err := ledgerClient.PostTransaction(rpcCtx, ledgerReq)
			if err != nil {
				// Log the error but don't fail the main operation (tolerate ledger being down)
				fmt.Printf("ledger post error: %v\n", err)
			} else if ack == nil || !ack.Ok {
				fmt.Printf("ledger ack negative: %v\n", ack)
			}
		}

		c.JSON(http.StatusOK, gin.H{"details": "Amount was sent"})

	})

}
