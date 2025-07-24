package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"grpc-server/internal/database/generated"
	"grpc-server/internal/models"
	"grpc-server/internal/repository"
)

// UserRepository implements repository.UserRepository using PostgreSQL
type UserRepository struct {
	conn    *pgx.Conn
	queries *database.Queries
}

// NewUserRepository creates a new PostgreSQL user repository
func NewUserRepository(conn *pgx.Conn) repository.UserRepository {
	return &UserRepository{
		conn:    conn,
		queries: database.New(conn),
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
	params, err := r.fromDomainUser(user)
	if err != nil {
		return err
	}

	dbUser, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		// Check for unique constraint violation (email already exists)
		if err.Error() == `ERROR: duplicate key value violates unique constraint "users_email_key" (SQLSTATE 23505)` {
			return repository.ErrEmailExists
		}
		return err
	}

	// Update the user with database-generated timestamps
	*user = *r.toDomainUser(dbUser)
	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	userUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, repository.ErrUserNotFound
	}

	var pgUUID pgtype.UUID
	err = pgUUID.Scan(userUUID)
	if err != nil {
		return nil, repository.ErrUserNotFound
	}

	dbUser, err := r.queries.GetUserByID(ctx, pgUUID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	return r.toDomainUser(dbUser), nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	dbUser, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	return r.toDomainUser(dbUser), nil
}

// Update modifies an existing user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	userUUID, err := uuid.Parse(user.ID)
	if err != nil {
		return repository.ErrUserNotFound
	}

	var id pgtype.UUID
	err = id.Scan(userUUID)
	if err != nil {
		return repository.ErrUserNotFound
	}

	var updatedAt pgtype.Timestamptz
	err = updatedAt.Scan(time.Now())
	if err != nil {
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
			return repository.ErrUserNotFound
		}
		// Check for unique constraint violation (email already exists)
		if err.Error() == `ERROR: duplicate key value violates unique constraint "users_email_key" (SQLSTATE 23505)` {
			return repository.ErrEmailExists
		}
		return err
	}

	// Update the user with database values
	*user = *r.toDomainUser(dbUser)
	return nil
}

// Delete removes a user by ID
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	userUUID, err := uuid.Parse(id)
	if err != nil {
		return repository.ErrUserNotFound
	}

	var pgUUID pgtype.UUID
	err = pgUUID.Scan(userUUID)
	if err != nil {
		return repository.ErrUserNotFound
	}

	err = r.queries.DeleteUser(ctx, pgUUID)
	if err != nil {
		return repository.ErrUserNotFound
	}

	return nil
}

// List returns paginated list of users sorted by creation time
func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]*models.User, int, error) {
	// Get total count
	totalCount, err := r.queries.CountUsers(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated users
	params := database.ListUsersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	dbUsers, err := r.queries.ListUsers(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	// Convert to domain models
	users := make([]*models.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = r.toDomainUser(dbUser)
	}

	return users, int(totalCount), nil
}

// EmailExists checks if email exists for a different user
func (r *UserRepository) EmailExists(ctx context.Context, email string, excludeID string) (bool, error) {
	userUUID, err := uuid.Parse(excludeID)
	if err != nil {
		return false, err
	}

	var pgUUID pgtype.UUID
	err = pgUUID.Scan(userUUID)
	if err != nil {
		return false, err
	}

	params := database.CheckEmailExistsParams{
		Email: email,
		ID:    pgUUID,
	}

	exists, err := r.queries.CheckEmailExists(ctx, params)
	if err != nil {
		return false, err
	}

	return exists, nil
}
