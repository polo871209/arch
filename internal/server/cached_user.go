package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"grpc-server/internal/cache"
	"grpc-server/internal/models"
	"grpc-server/internal/repository"
	"grpc-server/internal/validation"
	pb "grpc-server/pkg/pb"
)

// CachedUserServer implements the UserService with caching
type CachedUserServer struct {
	pb.UnimplementedUserServiceServer
	repo   repository.UserRepository
	cache  cache.Cache
	logger *slog.Logger
}

// NewCachedUserServer creates a new CachedUserServer instance
func NewCachedUserServer(repo repository.UserRepository, cache cache.Cache, logger *slog.Logger) *CachedUserServer {
	return &CachedUserServer{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

const (
	userCachePrefix     = "user:"
	userListCachePrefix = "users:list:"
	defaultCacheTTL     = 15 * time.Minute
)

func (s *CachedUserServer) userCacheKey(id string) string {
	return userCachePrefix + id
}

func (s *CachedUserServer) userListCacheKey(offset, limit int) string {
	return fmt.Sprintf("%s%d:%d", userListCachePrefix, offset, limit)
}

// CreateUser creates a new user and invalidates related cache
func (s *CachedUserServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
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
		s.logger.Error("Failed to create user", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create user")
	}

	// Cache the new user
	if err := s.cacheUser(ctx, user); err != nil {
		s.logger.Warn("Failed to cache new user", "id", user.ID, "error", err)
	}

	// Invalidate list cache
	s.invalidateListCache(ctx)

	s.logger.Info("User created", "id", user.ID, "email", user.Email)

	return &pb.CreateUserResponse{
		User:    user.ToProto(),
		Message: "User created successfully",
	}, nil
}

// GetUser retrieves a user by ID with caching
func (s *CachedUserServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	if err := validation.ValidateUserID(req.Id); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	// Try cache first
	cacheKey := s.userCacheKey(req.Id)
	cachedData, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		var user models.User
		if err := json.Unmarshal(cachedData, &user); err == nil {
			s.logger.Debug("Cache hit for user", "id", req.Id)
			return &pb.GetUserResponse{
				User:    user.ToProto(),
				Message: "User retrieved successfully",
			}, nil
		}
		s.logger.Warn("Failed to unmarshal cached user", "id", req.Id, "error", err)
	} else if err != cache.ErrCacheMiss {
		s.logger.Warn("Cache get failed", "id", req.Id, "error", err)
	}

	// Cache miss - get from database
	user, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return nil, status.Errorf(codes.NotFound, "user with ID %s not found", req.Id)
		}
		s.logger.Error("Failed to get user", "id", req.Id, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve user")
	}

	// Cache the user
	if err := s.cacheUser(ctx, user); err != nil {
		s.logger.Warn("Failed to cache user", "id", req.Id, "error", err)
	}

	return &pb.GetUserResponse{
		User:    user.ToProto(),
		Message: "User retrieved successfully",
	}, nil
}

// UpdateUser updates an existing user and invalidates cache
func (s *CachedUserServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	if err := validation.ValidateUpdateUser(req.Id, req.Name, req.Email, req.Age); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	// Get existing user from database (not cache) to ensure consistency
	user, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return nil, status.Errorf(codes.NotFound, "user with ID %s not found", req.Id)
		}
		s.logger.Error("Failed to get user for update", "id", req.Id, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve user")
	}

	// Check email uniqueness if email is being updated
	if req.Email != "" && req.Email != user.Email {
		exists, err := s.repo.EmailExists(ctx, req.Email, req.Id)
		if err != nil {
			s.logger.Error("Failed to check email existence", "email", req.Email, "error", err)
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
		s.logger.Error("Failed to update user", "id", req.Id, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to update user")
	}

	// Update cache
	if err := s.cacheUser(ctx, user); err != nil {
		s.logger.Warn("Failed to update cache", "id", req.Id, "error", err)
	}

	// Invalidate list cache
	s.invalidateListCache(ctx)

	s.logger.Info("User updated", "id", user.ID, "email", user.Email)

	return &pb.UpdateUserResponse{
		User:    user.ToProto(),
		Message: "User updated successfully",
	}, nil
}

// DeleteUser deletes a user by ID and invalidates cache
func (s *CachedUserServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	if err := validation.ValidateUserID(req.Id); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	if err := s.repo.Delete(ctx, req.Id); err != nil {
		if err == repository.ErrUserNotFound {
			return nil, status.Errorf(codes.NotFound, "user with ID %s not found", req.Id)
		}
		s.logger.Error("Failed to delete user", "id", req.Id, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to delete user")
	}

	// Remove from cache
	cacheKey := s.userCacheKey(req.Id)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		s.logger.Warn("Failed to delete user from cache", "id", req.Id, "error", err)
	}

	// Invalidate list cache
	s.invalidateListCache(ctx)

	s.logger.Info("User deleted", "id", req.Id)

	return &pb.DeleteUserResponse{
		Message: "User deleted successfully",
	}, nil
}

// ListUsers returns a paginated list of users with caching
func (s *CachedUserServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	// Validate and normalize pagination parameters
	page := max(req.Page, 1)
	limit := min(max(req.Limit, 1), 100) // Between 1 and 100
	offset := (page - 1) * limit

	// Try cache first
	cacheKey := s.userListCacheKey(int(offset), int(limit))
	cachedData, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		var response pb.ListUsersResponse
		if err := json.Unmarshal(cachedData, &response); err == nil {
			s.logger.Debug("Cache hit for user list", "offset", offset, "limit", limit)
			return &response, nil
		}
		s.logger.Warn("Failed to unmarshal cached user list", "error", err)
	} else if err != cache.ErrCacheMiss {
		s.logger.Warn("Cache get failed for user list", "error", err)
	}

	// Cache miss - get from database
	users, total, err := s.repo.List(ctx, int(offset), int(limit))
	if err != nil {
		s.logger.Error("Failed to list users", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve users")
	}

	// Convert to protobuf messages
	pbUsers := make([]*pb.User, len(users))
	for i, user := range users {
		pbUsers[i] = user.ToProto()
	}

	response := &pb.ListUsersResponse{
		Users:   pbUsers,
		Total:   int32(total),
		Message: fmt.Sprintf("Retrieved %d users (page %d)", len(pbUsers), page),
	}

	// Cache the response
	if responseData, err := json.Marshal(response); err == nil {
		if err := s.cache.Set(ctx, cacheKey, responseData, defaultCacheTTL); err != nil {
			s.logger.Warn("Failed to cache user list", "error", err)
		}
	}

	return response, nil
}

// Helper methods

func (s *CachedUserServer) cacheUser(ctx context.Context, user *models.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	cacheKey := s.userCacheKey(user.ID)
	return s.cache.Set(ctx, cacheKey, data, defaultCacheTTL)
}

func (s *CachedUserServer) invalidateListCache(ctx context.Context) {
	// Simple approach: delete common list cache patterns
	// In a production system, you might use cache tags or patterns
	for offset := 0; offset < 1000; offset += 10 {
		for limit := 1; limit <= 100; limit += 10 {
			cacheKey := s.userListCacheKey(offset, limit)
			if err := s.cache.Delete(ctx, cacheKey); err != nil {
				s.logger.Debug("Failed to invalidate list cache", "key", cacheKey, "error", err)
			}
		}
	}
}
