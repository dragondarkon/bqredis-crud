package repository

import (
	"context"
)

// BaseRepository defines generic CRUD operations
type BaseRepository[T any] interface {
	// GetAll retrieves all entities with pagination
	GetAll(ctx context.Context, params PaginationParams) ([]T, error)

	// GetByID retrieves an entity by ID
	GetByID(ctx context.Context, id string) (T, error)

	// Create creates a new entity
	Create(ctx context.Context, entity T) error

	// Update updates an existing entity
	Update(ctx context.Context, entity T) error

	// Delete removes an entity
	Delete(ctx context.Context, id string) error
}

// BaseRepositoryImpl provides a base implementation of common repository functionality
type BaseRepositoryImpl[T any] struct {
	// Common fields and utilities can be added here
}

// ValidatePagination ensures pagination parameters are valid
func (b *BaseRepositoryImpl[T]) ValidatePagination(params *PaginationParams) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 10
	}
}

// CalculateOffset calculates the offset for pagination
func (b *BaseRepositoryImpl[T]) CalculateOffset(params PaginationParams) int {
	return (params.Page - 1) * params.PageSize
}

// ValidateID checks if an ID is valid
func (b *BaseRepositoryImpl[T]) ValidateID(id string) error {
	if id == "" {
		return ErrNotFound
	}
	return nil
}
