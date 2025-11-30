package server

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	dbpkg "github.com/kamaalg/pocketPay/db"
	"github.com/kamaalg/pocketPay/ledger_service/ledgerpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	ledgerpb.UnimplementedLedgerServer
	pool *pgxpool.Pool
}

func (s *server) PostTransaction(ctx context.Context, req *ledgerpb.Transaction) (*ledgerpb.Ack, error) {
	const insertSQL = `INSERT INTO ledger(id, account_email, amount, description, created_at) VALUES($1, $2, $3, $4, to_timestamp($5))`
	_, err := s.pool.Exec(ctx, insertSQL, req.Id, req.AccountEmail, req.Amount, req.Description, req.Timestamp)
	if err != nil {
		fmt.Printf("db error: %v\n", err)
		return &ledgerpb.Ack{Ok: false, Id: req.Id, Message: "failed to write ledger"}, nil
	}

	return &ledgerpb.Ack{Ok: true, Id: req.Id, Message: "recorded"}, nil
}

func main() {
	dbURL := os.Getenv("DB_url")
	if dbURL == "" {
		fmt.Println("DB_url env var not set")
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := dbpkg.OpenDBPool(ctx, dbURL)
	if err != nil {
		fmt.Printf("failed to open db: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		fmt.Printf("failed to listen: %v\n", err)
		os.Exit(1)
	}

	srv := grpc.NewServer()
	ledgerpb.RegisterLedgerServer(srv, &server{pool: pool})
	// Enable reflection for easy debugging (grpcurl etc.)
	reflection.Register(srv)

	fmt.Println("ledger gRPC server listening on :50051")
	if err := srv.Serve(lis); err != nil {
		fmt.Printf("server error: %v\n", err)
		os.Exit(1)
	}
}
