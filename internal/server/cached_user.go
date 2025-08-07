package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/trace"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	grpc_codes "google.golang.org/grpc/codes"
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
	tracer trace.Tracer
}

func NewCachedUserServer(repo repository.UserRepository, cache cache.Cache, logger *slog.Logger) *CachedUserServer {
	return &CachedUserServer{
		repo:   repo,
		cache:  cache,
		logger: logging.New(logger),
		tracer: otel.Tracer("rpc-server.rpc/server"),
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
	log := logging.WithTrace(ctx, s.logger)
	user := models.NewUser(uuid.New().String(), req.Name, req.Email, req.Age)
	log.Debug("Created domain user model", logging.UserID, user.ID, logging.UserEmail, user.Email)

	if err := s.repo.Create(ctx, user); err != nil {
		if err == repository.ErrEmailExists {
			log.Warn("CreateUser email already exists", logging.UserEmail, req.Email)
			return nil, status.Errorf(grpc_codes.AlreadyExists, "user with email %s already exists", req.Email)
		}
		log.Error("Failed to create user in repository", logging.Error, err, logging.UserEmail, req.Email)
		return nil, status.Errorf(grpc_codes.Internal, "failed to create user")
	}

	if err := s.cacheUser(ctx, user); err != nil {
		log.Warn("Failed to cache new user", logging.UserID, user.ID, logging.Error, err)
	}

	// Invalidate list cache
	s.invalidateListCache(ctx)

	return &pb.CreateUserResponse{
		User:    user.ToProto(),
		Message: "User created successfully",
	}, nil
}

func (s *CachedUserServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	log := logging.WithTrace(ctx, s.logger)
	log.Debug("GetUser request received", logging.UserID, req.Id)

	// Try cache first
	cacheKey := s.userCacheKey(req.Id)
	log.Debug("Attempting cache lookup", logging.UserID, req.Id, logging.CacheKey, cacheKey)
	cachedData, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		var user models.User
		if err := json.Unmarshal(cachedData, &user); err == nil {
			log.Debug("Cache hit for user", logging.UserID, req.Id)
			return &pb.GetUserResponse{
				User:    user.ToProto(),
				Message: "User retrieved successfully",
			}, nil
		}
		log.Warn("Failed to unmarshal cached user", logging.UserID, req.Id, logging.Error, err)
	} else if err != cache.ErrCacheMiss {
		log.Warn("Cache get failed", logging.UserID, req.Id, logging.Error, err)
	}

	// Cache miss - get from database
	log.Debug("Cache miss, fetching from database", logging.UserID, req.Id)
	user, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		if err == repository.ErrUserNotFound {
			log.Info("User not found", logging.UserID, req.Id)
			return nil, status.Errorf(grpc_codes.NotFound, "user with ID %s not found", req.Id)
		}
		log.Error("Failed to get user from repository", logging.UserID, req.Id, logging.Error, err)
		return nil, status.Errorf(grpc_codes.Internal, "failed to retrieve user")
	}

	// Cache the user
	if err := s.cacheUser(ctx, user); err != nil {
		log.Warn("Failed to cache user", logging.UserID, req.Id, logging.Error, err)
	}

	log.Debug("User retrieved successfully", logging.UserID, user.ID, logging.UserEmail, user.Email)
	return &pb.GetUserResponse{
		User:    user.ToProto(),
		Message: "User retrieved successfully",
	}, nil
}

func (s *CachedUserServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	log := logging.WithTrace(ctx, s.logger)
	log.Debug("UpdateUser request received", logging.UserID, req.Id, "name", req.Name, logging.UserEmail, req.Email, "age", req.Age)

	// Get existing user from database (not cache) to ensure consistency
	log.Debug("Fetching existing user from database", logging.UserID, req.Id)
	user, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		if err == repository.ErrUserNotFound {
			log.Info("User not found for update", logging.UserID, req.Id)
			return nil, status.Errorf(grpc_codes.NotFound, "user with ID %s not found", req.Id)
		}
		log.Error("Failed to get user for update from repository", logging.UserID, req.Id, logging.Error, err)
		return nil, status.Errorf(grpc_codes.Internal, "failed to retrieve user")
	}

	// Check email uniqueness if email is being updated
	if req.Email != "" && req.Email != user.Email {
		log.Debug("Checking email uniqueness", "new_email", req.Email, logging.UserID, req.Id)
		exists, err := s.repo.EmailExists(ctx, req.Email, req.Id)
		if err != nil {
			log.Error("Failed to check email existence", logging.UserEmail, req.Email, logging.Error, err)
			return nil, status.Errorf(grpc_codes.Internal, "failed to validate email")
		}
		if exists {
			log.Warn("Email already exists for different user", logging.UserEmail, req.Email, logging.UserID, req.Id)
			return nil, status.Errorf(grpc_codes.AlreadyExists, "user with email %s already exists", req.Email)
		}
	}

	// Update user
	oldEmail := user.Email
	user.Update(req.Name, req.Email, req.Age)
	log.Debug("User model updated", logging.UserID, user.ID, "old_email", oldEmail, "new_email", user.Email)

	// Save updated user
	if err := s.repo.Update(ctx, user); err != nil {
		log.Error("Failed to update user in repository", logging.UserID, req.Id, logging.Error, err)
		return nil, status.Errorf(grpc_codes.Internal, "failed to update user")
	}

	// Update cache
	if err := s.cacheUser(ctx, user); err != nil {
		log.Warn("Failed to update cache", logging.UserID, req.Id, logging.Error, err)
	}

	// Invalidate list cache
	s.invalidateListCache(ctx)

	log.Info("User updated successfully", logging.UserID, user.ID, logging.UserEmail, user.Email)

	return &pb.UpdateUserResponse{
		User:    user.ToProto(),
		Message: "User updated successfully",
	}, nil
}

func (s *CachedUserServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	log := logging.WithTrace(ctx, s.logger)
	log.Debug("DeleteUser request received", logging.UserID, req.Id)

	if err := s.repo.Delete(ctx, req.Id); err != nil {
		if err == repository.ErrUserNotFound {
			log.Info("User not found for deletion", logging.UserID, req.Id)
			return nil, status.Errorf(grpc_codes.NotFound, "user with ID %s not found", req.Id)
		}
		log.Error("Failed to delete user from repository", logging.UserID, req.Id, logging.Error, err)
		return nil, status.Errorf(grpc_codes.Internal, "failed to delete user")
	}

	// Remove from cache
	cacheKey := s.userCacheKey(req.Id)
	log.Debug("Removing user from cache", logging.UserID, req.Id, logging.CacheKey, cacheKey)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		log.Warn("Failed to delete user from cache", logging.UserID, req.Id, logging.Error, err)
	}

	// Invalidate list cache
	s.invalidateListCache(ctx)

	log.Info("User deleted successfully", logging.UserID, req.Id)

	return &pb.DeleteUserResponse{
		Message: "User deleted successfully",
	}, nil
}

func (s *CachedUserServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	log := logging.WithTrace(ctx, s.logger)
	log.Debug("ListUsers request received", "page", req.Page, "limit", req.Limit)

	// Validate and normalize pagination parameters
	page := max(req.Page, 1)
	limit := min(max(req.Limit, 1), 100) // Between 1 and 100
	offset := (page - 1) * limit

	log.Debug("Normalized pagination parameters", "page", page, "limit", limit, "offset", offset)

	// Try cache first
	cacheKey := s.userListCacheKey(int(offset), int(limit))
	log.Debug("Attempting cache lookup for user list", logging.CacheKey, cacheKey)
	cachedData, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		var response pb.ListUsersResponse
		if err := json.Unmarshal(cachedData, &response); err == nil {
			log.Debug("Cache hit for user list", "offset", offset, "limit", limit, "total", response.Total)
			return &response, nil
		}
		log.Warn("Failed to unmarshal cached user list", logging.Error, err)
	} else if err != cache.ErrCacheMiss {
		log.Warn("Cache get failed for user list", logging.Error, err)
	}

	// Cache miss - get from database
	log.Debug("Cache miss, fetching user list from database", "offset", offset, "limit", limit)
	users, total, err := s.repo.List(ctx, int(offset), int(limit))
	if err != nil {
		log.Error("Failed to list users from repository", logging.Error, err)
		return nil, status.Errorf(grpc_codes.Internal, "failed to retrieve users")
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
			log.Warn("Failed to cache user list", logging.Error, err)
		} else {
			log.Debug("Cached user list", logging.CacheKey, cacheKey, "ttl", defaultCacheTTL)
		}
	}

	log.Debug("User list retrieved successfully", "total_count", total, "returned_count", len(users), "page", page)
	return response, nil
}

func (s *CachedUserServer) cacheUser(ctx context.Context, user *models.User) error {
	log := logging.WithTrace(ctx, s.logger)
	log.Debug("Caching user", logging.UserID, user.ID, logging.UserEmail, user.Email)

	data, err := json.Marshal(user)
	if err != nil {
		log.Error("Failed to marshal user for caching", logging.UserID, user.ID, logging.Error, err)
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
	ctx, span := s.tracer.Start(ctx, "cache.invalidate_list",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("cache.operation", "invalidate_list"),
		),
	)
	defer span.End()

	log := logging.WithTrace(ctx, s.logger)
	log.Debug("Starting list cache invalidation")
	invalidatedCount := 0

	// Simple approach: delete common list cache patterns (most used combinations)
	commonLimits := []int{10, 20, 50, 100}
	for offset := 0; offset < 100; offset += 10 { // Only check first 10 pages
		for _, limit := range commonLimits {
			cacheKey := s.userListCacheKey(offset, limit)
			// Use untraced delete to avoid creating individual spans
			if tracedCache, ok := s.cache.(*cache.TracedCache); ok {
				if err := tracedCache.DeleteUntraced(ctx, cacheKey); err == nil {
					invalidatedCount++
				}
			} else {
				// Fallback to regular delete if not a traced cache
				if err := s.cache.Delete(ctx, cacheKey); err == nil {
					invalidatedCount++
				}
			}
		}
	}

	span.SetAttributes(attribute.Int("cache.invalidated_entries", invalidatedCount))
	log.Debug("List cache invalidation completed", "invalidated_entries", invalidatedCount)
}
