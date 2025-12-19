package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"final/ledger"
	ledgerv1 "final/ledger/ledger/v1"
	"google.golang.org/grpc"
)

func main() {
	addr := getenv("LEDGER_GRPC_ADDR", "0.0.0.0:50051")

	svc, closeFn, err := ledger.New(context.Background())
	if err != nil {
		fmt.Println("ledger init error:", err)
		os.Exit(1)
	}
	defer func() { _ = closeFn() }()

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("listen error:", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	ledgerv1.RegisterLedgerServiceServer(grpcServer, ledger.NewGRPCServer(svc))

	fmt.Println("Ledger gRPC started on", addr)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = ctx
		grpcServer.GracefulStop()
	}()

	if err := grpcServer.Serve(lis); err != nil {
		fmt.Println("serve error:", err)
		os.Exit(1)
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
