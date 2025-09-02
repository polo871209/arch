package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"grpc-server/internal/cache"
	"grpc-server/internal/config"
	"grpc-server/internal/database"
	"grpc-server/internal/logging"
	"grpc-server/internal/repository/postgres"
	"grpc-server/internal/server"
	"grpc-server/internal/tracing"
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
	// Note: ensure import "grpc-server/internal/logging" is present for the TraceContextHandler
	if cfg.Logger.Format == "text" {
		handler = logging.NewTraceContextHandler(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: cfg.Logger.Level,
		}))
	} else {
		handler = logging.NewTraceContextHandler(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: cfg.Logger.Level,
		}))
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Initialize OpenTelemetry tracing
	var tracingShutdown func(context.Context) error
	if cfg.Tracing.Enabled {
		var err error
		tracingShutdown, err = tracing.InitTracing(ctx, tracing.TracingConfig{
			ServiceName:    cfg.Tracing.ServiceName,
			ServiceVersion: cfg.Tracing.ServiceVersion,
			CollectorURL:   cfg.Tracing.CollectorURL,
			Enabled:        cfg.Tracing.Enabled,
		})
		if err != nil {
			slog.Error("Failed to initialize tracing", "error", err)
			os.Exit(1)
		}
		defer func() {
			if err := tracingShutdown(ctx); err != nil {
				slog.Error("Failed to shutdown tracing", "error", err)
			}
		}()
	}

	// Create listener
	address := fmt.Sprintf(":%s", cfg.Server.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		slog.Error("Failed to listen", "address", address, "error", err)
		os.Exit(1)
	}

	// Create gRPC server with configuration and tracing interceptors
	grpcOpts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(cfg.Server.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(cfg.Server.MaxSendMsgSize),
	}

	// Add tracing interceptors if enabled
	if cfg.Tracing.Enabled {
		grpcOpts = append(grpcOpts, grpc.StatsHandler(otelgrpc.NewServerHandler()))
	}

	grpcServer := grpc.NewServer(grpcOpts...)

	// Connect to PostgreSQL database (with tracing)
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
	userRepo := postgres.NewUserRepository(dbPool, logger)

	// Connect to Valkey cache
	slog.Info("Connecting to Valkey cache")
	valkeyCache, err := cache.NewValkeyCache(&cfg.Cache, logger)
	if err != nil {
		slog.Error("Failed to connect to cache", "error", err)
		os.Exit(1)
	}
	defer valkeyCache.Close()

	// Wrap cache with tracing if enabled
	cacheInterface := cache.Cache(valkeyCache)
	if cfg.Tracing.Enabled {
		cacheInterface = cache.NewTracedCache(valkeyCache, cfg.Tracing.ServiceName)
	}

	// Create and register the combined service (user + test)
	combinedService := server.NewCombinedServer(userRepo, cacheInterface, logger)
	pb.RegisterUserServiceServer(grpcServer, combinedService)

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
			"tracing_enabled", cfg.Tracing.Enabled,
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
