package postgres

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"grpc-server/internal/database/generated"
	"grpc-server/internal/models"
	"grpc-server/internal/repository"
)

// UserRepository implements repository.UserRepository using PostgreSQL
type UserRepository struct {
	conn    *pgx.Conn
	queries *database.Queries
	tracer  trace.Tracer
}

// NewUserRepository creates a new PostgreSQL user repository
func NewUserRepository(conn *pgx.Conn) repository.UserRepository {
	return &UserRepository{
		conn:    conn,
		queries: database.New(conn),
		tracer:  otel.Tracer("postgres.user.repository"),
	}
}

// Helper functions to convert between domain models and database models
func (r *UserRepository) toDomainUser(dbUser database.User) *models.User {
	// Convert pgtype.UUID to string
	var idStr string
	if dbUser.ID.Valid {
		// Convert the UUID bytes to a proper UUID and then to string
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
	// Convert uuid.UUID to [16]byte for pgtype.UUID
	id.Bytes = userUUID
	id.Valid = true

	var createdAt, updatedAt pgtype.Timestamptz
	err = createdAt.Scan(user.CreatedAt)
	if err != nil {
		return database.CreateUserParams{}, err
	}
	err = updatedAt.Scan(user.UpdatedAt)
	if err != nil {
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

// Create stores a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	slog.Debug("Creating user", "user_id", user.ID, "email", user.Email)

	ctx, span := r.tracer.Start(ctx, "user.repository.create",
		trace.WithAttributes(
			attribute.String("user.id", user.ID),
			attribute.String("user.email", user.Email),
		),
	)
	defer span.End()

	params, err := r.fromDomainUser(user)
	if err != nil {
		slog.Error("Failed to convert domain user to database params", "error", err, "user_id", user.ID)
		span.RecordError(err)
		return err
	}

	dbUser, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		slog.Error("Failed to create user in database", "error", err, "user_id", user.ID, "email", user.Email)
		span.RecordError(err)
		// Check for unique constraint violation (email already exists)
		if err.Error() == `ERROR: duplicate key value violates unique constraint "users_email_key" (SQLSTATE 23505)` {
			return repository.ErrEmailExists
		}
		return err
	}

	// Update the user with database-generated timestamps
	*user = *r.toDomainUser(dbUser)
	span.SetAttributes(attribute.String("user.created_at", user.CreatedAt.String()))

	slog.Info("User created successfully", "user_id", user.ID, "email", user.Email, "created_at", user.CreatedAt)
	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	slog.Debug("Getting user by ID", "user_id", id)

	ctx, span := r.tracer.Start(ctx, "user.repository.get_by_id",
		trace.WithAttributes(
			attribute.String("user.id", id),
		),
	)
	defer span.End()

	userUUID, err := uuid.Parse(id)
	if err != nil {
		slog.Error("Invalid user ID format", "error", err, "user_id", id)
		span.RecordError(err)
		return nil, repository.ErrUserNotFound
	}

	var pgUUID pgtype.UUID
	err = pgUUID.Scan(userUUID)
	if err != nil {
		slog.Error("Failed to scan UUID", "error", err, "user_id", id)
		span.RecordError(err)
		return nil, repository.ErrUserNotFound
	}

	dbUser, err := r.queries.GetUserByID(ctx, pgUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			slog.Debug("User not found", "user_id", id)
			return nil, repository.ErrUserNotFound
		}
		slog.Error("Failed to get user from database", "error", err, "user_id", id)
		span.RecordError(err)
		return nil, err
	}

	user := r.toDomainUser(dbUser)
	span.SetAttributes(attribute.String("user.email", user.Email))

	slog.Debug("User retrieved successfully", "user_id", user.ID, "email", user.Email)
	return user, nil
}

// Update modifies an existing user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	slog.Debug("Updating user", "user_id", user.ID, "email", user.Email)

	ctx, span := r.tracer.Start(ctx, "user.repository.update",
		trace.WithAttributes(
			attribute.String("user.id", user.ID),
			attribute.String("user.email", user.Email),
		),
	)
	defer span.End()

	userUUID, err := uuid.Parse(user.ID)
	if err != nil {
		slog.Error("Invalid user ID format", "error", err, "user_id", user.ID)
		span.RecordError(err)
		return repository.ErrUserNotFound
	}

	var id pgtype.UUID
	err = id.Scan(userUUID)
	if err != nil {
		slog.Error("Failed to scan UUID", "error", err, "user_id", user.ID)
		span.RecordError(err)
		return repository.ErrUserNotFound
	}

	var updatedAt pgtype.Timestamptz
	err = updatedAt.Scan(time.Now())
	if err != nil {
		slog.Error("Failed to scan timestamp", "error", err, "user_id", user.ID)
		span.RecordError(err)
		return err
	}

	params := database.UpdateUserParams{
		ID:        id,
		Name:      user.Name,
		Email:     user.Email,
		Age:       user.Age,
		UpdatedAt: updatedAt,
	}

	dbUser, err := r.queries.UpdateUser(ctx, params)
	if err != nil {
		if err == pgx.ErrNoRows {
			slog.Debug("User not found for update", "user_id", user.ID)
			return repository.ErrUserNotFound
		}
		// Check for unique constraint violation (email already exists)
		if err.Error() == `ERROR: duplicate key value violates unique constraint "users_email_key" (SQLSTATE 23505)` {
			slog.Error("Email already exists", "email", user.Email, "user_id", user.ID)
			span.RecordError(err)
			return repository.ErrEmailExists
		}
		slog.Error("Failed to update user in database", "error", err, "user_id", user.ID)
		span.RecordError(err)
		return err
	}

	// Update the user with database values
	*user = *r.toDomainUser(dbUser)

	slog.Info("User updated successfully", "user_id", user.ID, "email", user.Email, "updated_at", user.UpdatedAt)
	return nil
}

// Delete removes a user by ID
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	slog.Debug("Deleting user", "user_id", id)

	ctx, span := r.tracer.Start(ctx, "user.repository.delete",
		trace.WithAttributes(
			attribute.String("user.id", id),
		),
	)
	defer span.End()

	userUUID, err := uuid.Parse(id)
	if err != nil {
		slog.Error("Invalid user ID format", "error", err, "user_id", id)
		span.RecordError(err)
		return repository.ErrUserNotFound
	}

	var pgUUID pgtype.UUID
	err = pgUUID.Scan(userUUID)
	if err != nil {
		slog.Error("Failed to scan UUID", "error", err, "user_id", id)
		span.RecordError(err)
		return repository.ErrUserNotFound
	}

	err = r.queries.DeleteUser(ctx, pgUUID)
	if err != nil {
		slog.Error("Failed to delete user from database", "error", err, "user_id", id)
		span.RecordError(err)
		return repository.ErrUserNotFound
	}

	slog.Info("User deleted successfully", "user_id", id)
	return nil
}

// List returns paginated list of users sorted by creation time
func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]*models.User, int, error) {
	slog.Debug("Listing users", "offset", offset, "limit", limit)

	ctx, span := r.tracer.Start(ctx, "user.repository.list",
		trace.WithAttributes(
			attribute.Int("pagination.offset", offset),
			attribute.Int("pagination.limit", limit),
		),
	)
	defer span.End()

	// Get total count
	totalCount, err := r.queries.CountUsers(ctx)
	if err != nil {
		slog.Error("Failed to count users", "error", err)
		span.RecordError(err)
		return nil, 0, err
	}

	// Get paginated users
	params := database.ListUsersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	dbUsers, err := r.queries.ListUsers(ctx, params)
	if err != nil {
		slog.Error("Failed to list users from database", "error", err, "offset", offset, "limit", limit)
		span.RecordError(err)
		return nil, 0, err
	}

	// Convert to domain models
	users := make([]*models.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = r.toDomainUser(dbUser)
	}

	span.SetAttributes(
		attribute.Int("result.total_count", int(totalCount)),
		attribute.Int("result.returned_count", len(users)),
	)

	slog.Debug("Users retrieved successfully", "total_count", totalCount, "returned_count", len(users), "offset", offset, "limit", limit)
	return users, int(totalCount), nil
}

// EmailExists checks if email exists for a different user
func (r *UserRepository) EmailExists(ctx context.Context, email string, excludeID string) (bool, error) {
	slog.Debug("Checking email existence", "email", email, "exclude_id", excludeID)

	ctx, span := r.tracer.Start(ctx, "user.repository.email_exists",
		trace.WithAttributes(
			attribute.String("email", email),
			attribute.String("exclude_id", excludeID),
		),
	)
	defer span.End()

	userUUID, err := uuid.Parse(excludeID)
	if err != nil {
		slog.Error("Invalid exclude ID format", "error", err, "exclude_id", excludeID)
		span.RecordError(err)
		return false, err
	}

	var pgUUID pgtype.UUID
	err = pgUUID.Scan(userUUID)
	if err != nil {
		slog.Error("Failed to scan UUID", "error", err, "exclude_id", excludeID)
		span.RecordError(err)
		return false, err
	}

	params := database.CheckEmailExistsParams{
		Email: email,
		ID:    pgUUID,
	}

	exists, err := r.queries.CheckEmailExists(ctx, params)
	if err != nil {
		slog.Error("Failed to check email existence", "error", err, "email", email, "exclude_id", excludeID)
		span.RecordError(err)
		return false, err
	}

	span.SetAttributes(attribute.Bool("result.exists", exists))
	slog.Debug("Email existence check completed", "email", email, "exclude_id", excludeID, "exists", exists)
	return exists, nil
}
