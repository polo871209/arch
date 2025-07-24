package memory

import (
	"context"
	"sort"
	"sync"

	"grpc-server/internal/models"
	"grpc-server/internal/repository"
)

// UserRepository implements repository.UserRepository using in-memory storage
type UserRepository struct {
	users map[string]*models.User
	mutex sync.RWMutex
}

// NewUserRepository creates a new in-memory user repository
func NewUserRepository() repository.UserRepository {
	return &UserRepository{
		users: make(map[string]*models.User),
	}
}

// Create stores a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.users[user.ID]; exists {
		return repository.ErrUserExists
	}

	// Check if email already exists
	for _, existingUser := range r.users {
		if existingUser.Email == user.Email {
			return repository.ErrEmailExists
		}
	}

	r.users[user.ID] = user
	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, repository.ErrUserNotFound
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}

	return nil, repository.ErrUserNotFound
}

// Update modifies an existing user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.users[user.ID]; !exists {
		return repository.ErrUserNotFound
	}

	r.users[user.ID] = user
	return nil
}

// Delete removes a user by ID
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.users[id]; !exists {
		return repository.ErrUserNotFound
	}

	delete(r.users, id)
	return nil
}

// List returns paginated list of users sorted by creation time
func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]*models.User, int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Convert map to slice and sort by creation time (newest first)
	allUsers := make([]*models.User, 0, len(r.users))
	for _, user := range r.users {
		allUsers = append(allUsers, user)
	}

	sort.Slice(allUsers, func(i, j int) bool {
		return allUsers[i].CreatedAt.After(allUsers[j].CreatedAt)
	})

	total := len(allUsers)

	// Apply pagination
	if offset >= total {
		return []*models.User{}, total, nil
	}

	end := min(offset+limit, total)

	return allUsers[offset:end], total, nil
}

// EmailExists checks if email exists for a different user
func (r *UserRepository) EmailExists(ctx context.Context, email string, excludeID string) (bool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for id, user := range r.users {
		if id != excludeID && user.Email == email {
			return true, nil
		}
	}

	return false, nil
}
