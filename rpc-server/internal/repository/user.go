package repository

import (
	"context"
	"errors"

	"grpc-server/internal/models"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
	ErrEmailExists  = errors.New("email already exists")
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]*models.User, int, error)
	EmailExists(ctx context.Context, email string, excludeID string) (bool, error)
}
