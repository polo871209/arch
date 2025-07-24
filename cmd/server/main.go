package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"grpc-server/internal/config"
	"grpc-server/internal/repository/memory"
	"grpc-server/internal/server"
	pb "grpc-server/pkg/pb"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup structured logging
	var handler slog.Handler
	if cfg.Logger.Format == "text" {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: cfg.Logger.Level,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: cfg.Logger.Level,
		})
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Create listener
	address := fmt.Sprintf(":%s", cfg.Server.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		slog.Error("Failed to listen", "address", address, "error", err)
		os.Exit(1)
	}

	// Create gRPC server with configuration
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(cfg.Server.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(cfg.Server.MaxSendMsgSize),
	)

	// Initialize repository and server
	userRepo := memory.NewUserRepository()
	userServer := server.NewUserServer(userRepo)

	// Register services
	pb.RegisterUserServiceServer(grpcServer, userServer)

	// Enable reflection if configured
	if cfg.Server.EnableReflection {
		reflection.Register(grpcServer)
		slog.Info("gRPC reflection enabled")
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("Shutdown signal received, stopping server...")
		grpcServer.GracefulStop()
		cancel()
	}()

	slog.Info("gRPC server starting",
		"address", address,
		"max_recv_size", cfg.Server.MaxRecvMsgSize,
		"max_send_size", cfg.Server.MaxSendMsgSize,
		"reflection", cfg.Server.EnableReflection)

	// Start serving
	if err := grpcServer.Serve(listener); err != nil {
		slog.Error("Failed to serve gRPC server", "error", err)
		os.Exit(1)
	}

	<-ctx.Done()
	slog.Info("Server stopped")
}
