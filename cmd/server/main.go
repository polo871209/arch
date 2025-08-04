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

	"grpc-server/internal/cache"
	"grpc-server/internal/config"
	"grpc-server/internal/database"
	"grpc-server/internal/repository/postgres"
	"grpc-server/internal/server"
	pb "grpc-server/pkg/pb"
)

func main() {
	// Create context for the entire application
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	// Connect to PostgreSQL database
	slog.Info("Connecting to PostgreSQL database")
	dbPool, err := database.Connect(ctx, &cfg.Database)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if dbPool != nil {
			dbPool.Close()
		}
	}()

	// Create PostgreSQL repository
	userRepo := postgres.NewUserRepository(dbPool)

	// Connect to Valkey cache
	slog.Info("Connecting to Valkey cache")
	valkeyCache, err := cache.NewValkeyCache(&cfg.Cache, logger)
	if err != nil {
		slog.Error("Failed to connect to cache", "error", err)
		os.Exit(1)
	}
	defer valkeyCache.Close()

	// Create and register the cached user service
	userService := server.NewCachedUserServer(userRepo, valkeyCache, logger)
	pb.RegisterUserServiceServer(grpcServer, userService)

	// Enable reflection if configured
	if cfg.Server.EnableReflection {
		reflection.Register(grpcServer)
		slog.Info("gRPC reflection enabled")
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		slog.Info("gRPC server starting",
			"address", address,
			"max_recv_size", cfg.Server.MaxRecvMsgSize,
			"max_send_size", cfg.Server.MaxSendMsgSize,
			"reflection", cfg.Server.EnableReflection,
		)
		if err := grpcServer.Serve(listener); err != nil {
			slog.Error("gRPC server failed", "error", err)
			cancel()
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	slog.Info("Shutdown signal received, stopping server...")

	// Graceful shutdown
	grpcServer.GracefulStop()
	slog.Info("Server stopped gracefully")
}
