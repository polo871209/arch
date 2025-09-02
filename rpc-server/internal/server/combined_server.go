package server

import (
	"context"
	"log/slog"

	"grpc-server/internal/cache"
	"grpc-server/internal/repository"
	pb "grpc-server/pkg/pb"
)

type CombinedServer struct {
	pb.UnimplementedUserServiceServer
	cachedUserServer *CachedUserServer
	testServer       *TestServer
}

func NewCombinedServer(userRepo repository.UserRepository, cache cache.Cache, logger *slog.Logger) *CombinedServer {
	return &CombinedServer{
		cachedUserServer: NewCachedUserServer(userRepo, cache, logger),
		testServer:       NewTestServer(logger),
	}
}

// Implement all UserService methods by delegating to the appropriate server
func (s *CombinedServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	return s.cachedUserServer.CreateUser(ctx, req)
}

func (s *CombinedServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	return s.cachedUserServer.GetUser(ctx, req)
}

func (s *CombinedServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	return s.cachedUserServer.UpdateUser(ctx, req)
}

func (s *CombinedServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	return s.cachedUserServer.DeleteUser(ctx, req)
}

func (s *CombinedServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	return s.cachedUserServer.ListUsers(ctx, req)
}

func (s *CombinedServer) TestError(ctx context.Context, req *pb.TestErrorRequest) (*pb.TestErrorResponse, error) {
	return s.testServer.TestError(ctx, req)
}
