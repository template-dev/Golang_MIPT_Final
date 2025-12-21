package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"final/gateway/internal/middleware"
	"final/gateway/internal/server"
	ledgerv1 "final/gen/ledger/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	addr := getenv("LEDGER_ADDR", "localhost:50051")

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println("grpc dial error:", err)
		os.Exit(1)
	}
	defer conn.Close()

	client := ledgerv1.NewLedgerServiceClient(conn)

	h := server.NewRouter(client)

	handler := middleware.Timeout(h)
	handler = middleware.Logging(handler)

	fmt.Println("Gateway started on :8080, ledger:", addr)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		fmt.Println("http server error:", err)
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

var _ = context.Background
