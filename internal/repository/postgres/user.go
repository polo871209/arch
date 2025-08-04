package postgres

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	database "grpc-server/internal/database/generated"
	"grpc-server/internal/logging"
	"grpc-server/internal/models"
	"grpc-server/internal/repository"
)

type UserRepository struct {
	pool    *pgxpool.Pool
	queries *database.Queries
}

func NewUserRepository(pool *pgxpool.Pool) repository.UserRepository {
	return &UserRepository{
		pool:    pool,
		queries: database.New(pool),
	}
}

// Helper to parse UUID string to pgtype.UUID
func parseUUID(id string) (pgtype.UUID, error) {
	userUUID, err := uuid.Parse(id)
	if err != nil {
		return pgtype.UUID{}, err
	}

	return pgtype.UUID{
		Bytes: userUUID,
		Valid: true,
	}, nil
}

func (r *UserRepository) toDomainUser(dbUser database.User) *models.User {
	var idStr string
	if dbUser.ID.Valid {
		u := uuid.UUID(dbUser.ID.Bytes)
		idStr = u.String()
	}

	user := &models.User{
		ID:    idStr,
		Name:  dbUser.Name,
		Email: dbUser.Email,
		Age:   dbUser.Age,
	}

	if dbUser.CreatedAt.Valid {
		user.CreatedAt = dbUser.CreatedAt.Time
	}
	if dbUser.UpdatedAt.Valid {
		user.UpdatedAt = dbUser.UpdatedAt.Time
	}

	return user
}

func (r *UserRepository) fromDomainUser(user *models.User) (database.CreateUserParams, error) {
	userUUID, err := uuid.Parse(user.ID)
	if err != nil {
		return database.CreateUserParams{}, err
	}

	var id pgtype.UUID
	id.Bytes = userUUID
	id.Valid = true

	var createdAt, updatedAt pgtype.Timestamptz
	if err := createdAt.Scan(user.CreatedAt); err != nil {
		return database.CreateUserParams{}, err
	}
	if err := updatedAt.Scan(user.UpdatedAt); err != nil {
		return database.CreateUserParams{}, err
	}

	return database.CreateUserParams{
		ID:        id,
		Name:      user.Name,
		Email:     user.Email,
		Age:       user.Age,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	slog.Debug("Creating user", logging.UserID, user.ID, logging.UserEmail, user.Email)

	params, err := r.fromDomainUser(user)
	if err != nil {
		slog.Error("Failed to convert domain user to database params", logging.Error, err, logging.UserID, user.ID)
		return err
	}

	dbUser, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		slog.Error("Failed to create user in database", logging.Error, err, logging.UserID, user.ID, logging.UserEmail, user.Email)
		if err.Error() == `ERROR: duplicate key value violates unique constraint "users_email_key" (SQLSTATE 23505)` {
			return repository.ErrEmailExists
		}
		return err
	}

	*user = *r.toDomainUser(dbUser)

	slog.Info("User created successfully", logging.UserID, user.ID, logging.UserEmail, user.Email, "created_at", user.CreatedAt)
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	slog.Debug("Getting user by ID", logging.UserID, id)

	pgUUID, err := parseUUID(id)
	if err != nil {
		slog.Error("Invalid user ID format", logging.Error, err, logging.UserID, id)
		return nil, repository.ErrUserNotFound
	}

	dbUser, err := r.queries.GetUserByID(ctx, pgUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			slog.Debug("User not found", logging.UserID, id)
			return nil, repository.ErrUserNotFound
		}
		slog.Error("Failed to get user from database", logging.Error, err, logging.UserID, id)
		return nil, err
	}

	user := r.toDomainUser(dbUser)

	slog.Debug("User retrieved successfully", logging.UserID, user.ID, logging.UserEmail, user.Email)
	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	slog.Debug("Updating user", logging.UserID, user.ID, logging.UserEmail, user.Email)

	pgUUID, err := parseUUID(user.ID)
	if err != nil {
		slog.Error("Invalid user ID format", logging.Error, err, logging.UserID, user.ID)
		return repository.ErrUserNotFound
	}

	var updatedAt pgtype.Timestamptz
	if err := updatedAt.Scan(time.Now()); err != nil {
		slog.Error("Failed to scan timestamp", logging.Error, err, logging.UserID, user.ID)
		return err
	}

	params := database.UpdateUserParams{
		ID:        pgUUID,
		Name:      user.Name,
		Email:     user.Email,
		Age:       user.Age,
		UpdatedAt: updatedAt,
	}

	dbUser, err := r.queries.UpdateUser(ctx, params)
	if err != nil {
		if err == pgx.ErrNoRows {
			slog.Debug("User not found for update", logging.UserID, user.ID)
			return repository.ErrUserNotFound
		}
		if err.Error() == `ERROR: duplicate key value violates unique constraint "users_email_key" (SQLSTATE 23505)` {
			slog.Error("Email already exists", logging.UserEmail, user.Email, logging.UserID, user.ID)
			return repository.ErrEmailExists
		}
		slog.Error("Failed to update user in database", logging.Error, err, logging.UserID, user.ID)
		return err
	}

	*user = *r.toDomainUser(dbUser)

	slog.Info("User updated successfully", logging.UserID, user.ID, logging.UserEmail, user.Email, "updated_at", user.UpdatedAt)
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	slog.Debug("Deleting user", logging.UserID, id)

	pgUUID, err := parseUUID(id)
	if err != nil {
		slog.Error("Invalid user ID format", logging.Error, err, logging.UserID, id)
		return repository.ErrUserNotFound
	}

	if err := r.queries.DeleteUser(ctx, pgUUID); err != nil {
		slog.Error("Failed to delete user from database", logging.Error, err, logging.UserID, id)
		return repository.ErrUserNotFound
	}

	slog.Info("User deleted successfully", logging.UserID, id)
	return nil
}

func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]*models.User, int, error) {
	slog.Debug("Listing users", "offset", offset, "limit", limit)

	totalCount, err := r.queries.CountUsers(ctx)
	if err != nil {
		slog.Error("Failed to count users", logging.Error, err)
		return nil, 0, err
	}

	params := database.ListUsersParams{Limit: int32(limit), Offset: int32(offset)}
	dbUsers, err := r.queries.ListUsers(ctx, params)
	if err != nil {
		slog.Error("Failed to list users from database", logging.Error, err, "offset", offset, "limit", limit)
		return nil, 0, err
	}

	users := make([]*models.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = r.toDomainUser(dbUser)
	}

	slog.Debug("Users retrieved successfully", "total_count", totalCount, "returned_count", len(users), "offset", offset, "limit", limit)

	return users, int(totalCount), nil
}

func (r *UserRepository) EmailExists(ctx context.Context, email string, excludeID string) (bool, error) {
	slog.Debug("Checking email existence", logging.UserEmail, email, "exclude_id", excludeID)

	pgUUID, err := parseUUID(excludeID)
	if err != nil {
		slog.Error("Invalid exclude ID format", logging.Error, err, "exclude_id", excludeID)
		return false, err
	}

	params := database.CheckEmailExistsParams{Email: email, ID: pgUUID}
	exists, err := r.queries.CheckEmailExists(ctx, params)
	if err != nil {
		slog.Error("Failed to check email existence", logging.Error, err, logging.UserEmail, email, "exclude_id", excludeID)
		return false, err
	}

	slog.Debug("Email existence check completed", logging.UserEmail, email, "exclude_id", excludeID, "exists", exists)
	return exists, nil
}
