package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"grpc-server/internal/models"
	"grpc-server/internal/repository"
	"grpc-server/internal/validation"
	pb "grpc-server/pkg/pb"
)

// UserServer implements the UserService with clean architecture
type UserServer struct {
	pb.UnimplementedUserServiceServer
	repo repository.UserRepository
}

// NewUserServer creates a new UserServer instance
func NewUserServer(repo repository.UserRepository) *UserServer {
	return &UserServer{
		repo: repo,
	}
}

// CreateUser creates a new user with enhanced validation
func (s *UserServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	// Validate input
	if err := validation.ValidateCreateUser(req.Name, req.Email, req.Age); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	// Create domain model
	user := models.NewUser(uuid.New().String(), req.Name, req.Email, req.Age)

	// Store user
	if err := s.repo.Create(ctx, user); err != nil {
		if err == repository.ErrEmailExists {
			return nil, status.Errorf(codes.AlreadyExists, "user with email %s already exists", req.Email)
		}
		slog.Error("Failed to create user", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create user")
	}

	slog.Info("User created", "id", user.ID, "email", user.Email)

	return &pb.CreateUserResponse{
		User:    user.ToProto(),
		Message: "User created successfully",
	}, nil
}

// GetUser retrieves a user by ID
func (s *UserServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	if err := validation.ValidateUserID(req.Id); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	user, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return nil, status.Errorf(codes.NotFound, "user with ID %s not found", req.Id)
		}
		slog.Error("Failed to get user", "id", req.Id, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve user")
	}

	return &pb.GetUserResponse{
		User:    user.ToProto(),
		Message: "User retrieved successfully",
	}, nil
}

// UpdateUser updates an existing user with improved validation
func (s *UserServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	if err := validation.ValidateUpdateUser(req.Id, req.Name, req.Email, req.Age); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	// Get existing user
	user, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return nil, status.Errorf(codes.NotFound, "user with ID %s not found", req.Id)
		}
		slog.Error("Failed to get user for update", "id", req.Id, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve user")
	}

	// Check email uniqueness if email is being updated
	if req.Email != "" && req.Email != user.Email {
		exists, err := s.repo.EmailExists(ctx, req.Email, req.Id)
		if err != nil {
			slog.Error("Failed to check email existence", "email", req.Email, "error", err)
			return nil, status.Errorf(codes.Internal, "failed to validate email")
		}
		if exists {
			return nil, status.Errorf(codes.AlreadyExists, "user with email %s already exists", req.Email)
		}
	}

	// Update user
	user.Update(req.Name, req.Email, req.Age)

	// Save updated user
	if err := s.repo.Update(ctx, user); err != nil {
		slog.Error("Failed to update user", "id", req.Id, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to update user")
	}

	slog.Info("User updated", "id", user.ID, "email", user.Email)

	return &pb.UpdateUserResponse{
		User:    user.ToProto(),
		Message: "User updated successfully",
	}, nil
}

// DeleteUser deletes a user by ID
func (s *UserServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	if err := validation.ValidateUserID(req.Id); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	if err := s.repo.Delete(ctx, req.Id); err != nil {
		if err == repository.ErrUserNotFound {
			return nil, status.Errorf(codes.NotFound, "user with ID %s not found", req.Id)
		}
		slog.Error("Failed to delete user", "id", req.Id, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to delete user")
	}

	slog.Info("User deleted", "id", req.Id)

	return &pb.DeleteUserResponse{
		Message: "User deleted successfully",
	}, nil
}

// ListUsers returns a paginated list of users with improved sorting
func (s *UserServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	// Validate and normalize pagination parameters
	page := max(req.Page, 1)
	limit := min(max(req.Limit, 1), 100) // Between 1 and 100
	offset := (page - 1) * limit

	users, total, err := s.repo.List(ctx, int(offset), int(limit))
	if err != nil {
		slog.Error("Failed to list users", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve users")
	}

	// Convert to protobuf messages
	pbUsers := make([]*pb.User, len(users))
	for i, user := range users {
		pbUsers[i] = user.ToProto()
	}

	return &pb.ListUsersResponse{
		Users:   pbUsers,
		Total:   int32(total),
		Message: fmt.Sprintf("Retrieved %d users (page %d)", len(pbUsers), page),
	}, nil
}
