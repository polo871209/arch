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
	"grpc-server/internal/logging"
	"grpc-server/internal/models"
	"grpc-server/internal/repository"
	pb "grpc-server/pkg/pb"
)

type CachedUserServer struct {
	pb.UnimplementedUserServiceServer
	repo   repository.UserRepository
	cache  cache.Cache
	logger *logging.Logger
}

func NewCachedUserServer(repo repository.UserRepository, cache cache.Cache, logger *slog.Logger) *CachedUserServer {
	return &CachedUserServer{
		repo:   repo,
		cache:  cache,
		logger: logging.New(logger),
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

func (s *CachedUserServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	user := models.NewUser(uuid.New().String(), req.Name, req.Email, req.Age)
	s.logger.Debug("Created domain user model", logging.UserID, user.ID, logging.UserEmail, user.Email)

	if err := s.repo.Create(ctx, user); err != nil {
		if err == repository.ErrEmailExists {
			s.logger.Warn("CreateUser email already exists", logging.UserEmail, req.Email)
			return nil, status.Errorf(codes.AlreadyExists, "user with email %s already exists", req.Email)
		}
		s.logger.Error("Failed to create user in repository", logging.Error, err, logging.UserEmail, req.Email)
		return nil, status.Errorf(codes.Internal, "failed to create user")
	}

	if err := s.cacheUser(ctx, user); err != nil {
		s.logger.Warn("Failed to cache new user", logging.UserID, user.ID, logging.Error, err)
	}

	// Invalidate list cache
	s.invalidateListCache(ctx)

	return &pb.CreateUserResponse{
		User:    user.ToProto(),
		Message: "User created successfully",
	}, nil
}

func (s *CachedUserServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	s.logger.Debug("GetUser request received", logging.UserID, req.Id)

	// Try cache first
	cacheKey := s.userCacheKey(req.Id)
	s.logger.Debug("Attempting cache lookup", logging.UserID, req.Id, logging.CacheKey, cacheKey)
	cachedData, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		var user models.User
		if err := json.Unmarshal(cachedData, &user); err == nil {
			s.logger.Debug("Cache hit for user", logging.UserID, req.Id)
			return &pb.GetUserResponse{
				User:    user.ToProto(),
				Message: "User retrieved successfully",
			}, nil
		}
		s.logger.Warn("Failed to unmarshal cached user", logging.UserID, req.Id, logging.Error, err)
	} else if err != cache.ErrCacheMiss {
		s.logger.Warn("Cache get failed", logging.UserID, req.Id, logging.Error, err)
	}

	// Cache miss - get from database
	s.logger.Debug("Cache miss, fetching from database", logging.UserID, req.Id)
	user, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		if err == repository.ErrUserNotFound {
			s.logger.Info("User not found", logging.UserID, req.Id)
			return nil, status.Errorf(codes.NotFound, "user with ID %s not found", req.Id)
		}
		s.logger.Error("Failed to get user from repository", logging.UserID, req.Id, logging.Error, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve user")
	}

	// Cache the user
	if err := s.cacheUser(ctx, user); err != nil {
		s.logger.Warn("Failed to cache user", logging.UserID, req.Id, logging.Error, err)
	}

	s.logger.Debug("User retrieved successfully", logging.UserID, user.ID, logging.UserEmail, user.Email)
	return &pb.GetUserResponse{
		User:    user.ToProto(),
		Message: "User retrieved successfully",
	}, nil
}

func (s *CachedUserServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	s.logger.Debug("UpdateUser request received", logging.UserID, req.Id, "name", req.Name, logging.UserEmail, req.Email, "age", req.Age)

	// Get existing user from database (not cache) to ensure consistency
	s.logger.Debug("Fetching existing user from database", logging.UserID, req.Id)
	user, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		if err == repository.ErrUserNotFound {
			s.logger.Info("User not found for update", logging.UserID, req.Id)
			return nil, status.Errorf(codes.NotFound, "user with ID %s not found", req.Id)
		}
		s.logger.Error("Failed to get user for update from repository", logging.UserID, req.Id, logging.Error, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve user")
	}

	// Check email uniqueness if email is being updated
	if req.Email != "" && req.Email != user.Email {
		s.logger.Debug("Checking email uniqueness", "new_email", req.Email, logging.UserID, req.Id)
		exists, err := s.repo.EmailExists(ctx, req.Email, req.Id)
		if err != nil {
			s.logger.Error("Failed to check email existence", logging.UserEmail, req.Email, logging.Error, err)
			return nil, status.Errorf(codes.Internal, "failed to validate email")
		}
		if exists {
			s.logger.Warn("Email already exists for different user", logging.UserEmail, req.Email, logging.UserID, req.Id)
			return nil, status.Errorf(codes.AlreadyExists, "user with email %s already exists", req.Email)
		}
	}

	// Update user
	oldEmail := user.Email
	user.Update(req.Name, req.Email, req.Age)
	s.logger.Debug("User model updated", logging.UserID, user.ID, "old_email", oldEmail, "new_email", user.Email)

	// Save updated user
	if err := s.repo.Update(ctx, user); err != nil {
		s.logger.Error("Failed to update user in repository", logging.UserID, req.Id, logging.Error, err)
		return nil, status.Errorf(codes.Internal, "failed to update user")
	}

	// Update cache
	if err := s.cacheUser(ctx, user); err != nil {
		s.logger.Warn("Failed to update cache", logging.UserID, req.Id, logging.Error, err)
	}

	// Invalidate list cache
	s.invalidateListCache(ctx)

	s.logger.Info("User updated successfully", logging.UserID, user.ID, logging.UserEmail, user.Email)

	return &pb.UpdateUserResponse{
		User:    user.ToProto(),
		Message: "User updated successfully",
	}, nil
}

func (s *CachedUserServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	s.logger.Debug("DeleteUser request received", logging.UserID, req.Id)

	if err := s.repo.Delete(ctx, req.Id); err != nil {
		if err == repository.ErrUserNotFound {
			s.logger.Info("User not found for deletion", logging.UserID, req.Id)
			return nil, status.Errorf(codes.NotFound, "user with ID %s not found", req.Id)
		}
		s.logger.Error("Failed to delete user from repository", logging.UserID, req.Id, logging.Error, err)
		return nil, status.Errorf(codes.Internal, "failed to delete user")
	}

	// Remove from cache
	cacheKey := s.userCacheKey(req.Id)
	s.logger.Debug("Removing user from cache", logging.UserID, req.Id, logging.CacheKey, cacheKey)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		s.logger.Warn("Failed to delete user from cache", logging.UserID, req.Id, logging.Error, err)
	}

	// Invalidate list cache
	s.invalidateListCache(ctx)

	s.logger.Info("User deleted successfully", logging.UserID, req.Id)

	return &pb.DeleteUserResponse{
		Message: "User deleted successfully",
	}, nil
}

func (s *CachedUserServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	s.logger.Debug("ListUsers request received", "page", req.Page, "limit", req.Limit)

	// Validate and normalize pagination parameters
	page := max(req.Page, 1)
	limit := min(max(req.Limit, 1), 100) // Between 1 and 100
	offset := (page - 1) * limit

	s.logger.Debug("Normalized pagination parameters", "page", page, "limit", limit, "offset", offset)

	// Try cache first
	cacheKey := s.userListCacheKey(int(offset), int(limit))
	s.logger.Debug("Attempting cache lookup for user list", logging.CacheKey, cacheKey)
	cachedData, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		var response pb.ListUsersResponse
		if err := json.Unmarshal(cachedData, &response); err == nil {
			s.logger.Debug("Cache hit for user list", "offset", offset, "limit", limit, "total", response.Total)
			return &response, nil
		}
		s.logger.Warn("Failed to unmarshal cached user list", logging.Error, err)
	} else if err != cache.ErrCacheMiss {
		s.logger.Warn("Cache get failed for user list", logging.Error, err)
	}

	// Cache miss - get from database
	s.logger.Debug("Cache miss, fetching user list from database", "offset", offset, "limit", limit)
	users, total, err := s.repo.List(ctx, int(offset), int(limit))
	if err != nil {
		s.logger.Error("Failed to list users from repository", logging.Error, err)
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
			s.logger.Warn("Failed to cache user list", logging.Error, err)
		} else {
			s.logger.Debug("Cached user list", logging.CacheKey, cacheKey, "ttl", defaultCacheTTL)
		}
	}

	s.logger.Debug("User list retrieved successfully", "total_count", total, "returned_count", len(users), "page", page)
	return response, nil
}

func (s *CachedUserServer) cacheUser(ctx context.Context, user *models.User) error {
	s.logger.Debug("Caching user", logging.UserID, user.ID, logging.UserEmail, user.Email)

	data, err := json.Marshal(user)
	if err != nil {
		s.logger.Error("Failed to marshal user for caching", logging.UserID, user.ID, logging.Error, err)
		return err
	}

	cacheKey := s.userCacheKey(user.ID)
	err = s.cache.Set(ctx, cacheKey, data, defaultCacheTTL)
	if err != nil {
		s.logger.Error("Failed to set user in cache", logging.UserID, user.ID, logging.CacheKey, cacheKey, logging.Error, err)
		return err
	}

	s.logger.Debug("User cached successfully", logging.UserID, user.ID, logging.CacheKey, cacheKey, "ttl", defaultCacheTTL)
	return nil
}

func (s *CachedUserServer) invalidateListCache(ctx context.Context) {
	s.logger.Debug("Starting list cache invalidation")
	invalidatedCount := 0

	// Simple approach: delete common list cache patterns (most used combinations)
	commonLimits := []int{10, 20, 50, 100}
	for offset := 0; offset < 100; offset += 10 { // Only check first 10 pages
		for _, limit := range commonLimits {
			cacheKey := s.userListCacheKey(offset, limit)
			if err := s.cache.Delete(ctx, cacheKey); err == nil {
				invalidatedCount++
			}
		}
	}

	s.logger.Debug("List cache invalidation completed", "invalidated_entries", invalidatedCount)
}
