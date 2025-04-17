package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dragondarkon/bqredis-crud/internal/domain/entity"
	"github.com/dragondarkon/bqredis-crud/internal/repository"
	"github.com/google/uuid"
)

// Custom error types
var (
	ErrUserNotFound = errors.New("user not found")
	ErrValidation   = errors.New("validation error")
)

// UserUseCase implements the business logic for user operations
type UserUseCase struct {
	primaryRepo repository.UserRepository
	cacheRepo   repository.UserRepository
}

// validateUser validates user fields
func (uc *UserUseCase) validateUser(user *entity.User, isCreate bool) error {
	if !isCreate && user.ID == "" {
		return fmt.Errorf("%w: id is required", ErrValidation)
	}
	if user.Name == "" {
		return fmt.Errorf("%w: name is required", ErrValidation)
	}
	if user.Email == "" {
		return fmt.Errorf("%w: email is required", ErrValidation)
	}
	return nil
}

// NewUserUseCase creates a new user use case
func NewUserUseCase(primaryRepo, cacheRepo repository.UserRepository) *UserUseCase {
	return &UserUseCase{
		primaryRepo: primaryRepo,
		cacheRepo:   cacheRepo,
	}
}

// GetAllUsers retrieves all users with pagination
func (uc *UserUseCase) GetAllUsers(ctx context.Context, page, pageSize int) ([]entity.User, error) {
	params := repository.PaginationParams{
		Page:     max(page, 1),
		PageSize: max(pageSize, 10),
	}

	// Use cache repository which handles caching internally
	users, err := uc.cacheRepo.GetAll(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	return users, nil
}

// GetUserByID retrieves a user by ID
func (uc *UserUseCase) GetUserByID(ctx context.Context, id string) (entity.User, error) {
	if id == "" {
		return entity.User{}, fmt.Errorf("%w: id is required", ErrValidation)
	}

	// Use cache repository which handles caching internally
	user, err := uc.cacheRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.User{}, ErrUserNotFound
		}
		return entity.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// CreateUser creates a new user
func (uc *UserUseCase) CreateUser(ctx context.Context, user entity.User) (entity.User, error) {
	if err := uc.validateUser(&user, true); err != nil {
		return entity.User{}, err
	}

	// Set ID and timestamps
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Use cache repository which handles cache invalidation internally
	if err := uc.cacheRepo.Create(ctx, user); err != nil {
		return entity.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// UpdateUser updates an existing user
func (uc *UserUseCase) UpdateUser(ctx context.Context, user entity.User) (entity.User, error) {
	if err := uc.validateUser(&user, false); err != nil {
		return entity.User{}, err
	}

	user.UpdatedAt = time.Now()

	// Use cache repository which handles cache invalidation internally
	if err := uc.cacheRepo.Update(ctx, user); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.User{}, ErrUserNotFound
		}
		return entity.User{}, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// DeleteUser removes a user
func (uc *UserUseCase) DeleteUser(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: id is required", ErrValidation)
	}

	// Use cache repository which handles cache invalidation internally
	if err := uc.cacheRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// Helper function for max value
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
