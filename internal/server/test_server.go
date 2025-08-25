package server

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	grpc_codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"grpc-server/internal/logging"
	pb "grpc-server/pkg/pb"
)

type TestServer struct {
	pb.UnimplementedUserServiceServer
	logger *logging.Logger
	tracer trace.Tracer
}

func NewTestServer(logger *slog.Logger) *TestServer {
	return &TestServer{
		logger: logging.New(logger),
		tracer: otel.Tracer("rpc-server.rpc/server"),
	}
}

func (s *TestServer) TestError(ctx context.Context, req *pb.TestErrorRequest) (*pb.TestErrorResponse, error) {
	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID().String()

	statusCode, err := strconv.Atoi(req.StatusCode)
	if err != nil || statusCode < 100 || statusCode > 599 {
		statusCode = 500 // Default to 500 if invalid input
	}

	s.logger.InfoCtx(ctx, "TestError request received", "status_code", statusCode, "trace_id", traceID)

	// Map HTTP status codes to gRPC codes
	var grpcCode grpc_codes.Code
	var message string

	switch statusCode {
	case 400:
		grpcCode = grpc_codes.InvalidArgument
		message = "Bad Request"
	case 401:
		grpcCode = grpc_codes.Unauthenticated
		message = "Unauthorized"
	case 403:
		grpcCode = grpc_codes.PermissionDenied
		message = "Forbidden"
	case 404:
		grpcCode = grpc_codes.NotFound
		message = "Not Found"
	case 409:
		grpcCode = grpc_codes.AlreadyExists
		message = "Conflict"
	case 429:
		grpcCode = grpc_codes.ResourceExhausted
		message = "Too Many Requests"
	case 500:
		grpcCode = grpc_codes.Internal
		message = "Internal Server Error"
	case 501:
		grpcCode = grpc_codes.Unimplemented
		message = "Not Implemented"
	case 503:
		grpcCode = grpc_codes.Unavailable
		message = "Service Unavailable"
	case 504:
		grpcCode = grpc_codes.DeadlineExceeded
		message = "Gateway Timeout"
	default:
		grpcCode = grpc_codes.Internal
		message = fmt.Sprintf("Test Error - Status %d", statusCode)
	}

	return nil, status.Errorf(grpcCode, "%s (trace_id: %s)", message, traceID)
}
