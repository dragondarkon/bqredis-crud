package repository

import (
	"errors"

	"github.com/dragondarkon/bqredis-crud/internal/domain/entity"
)

// Common repository errors
var (
	ErrNotFound = errors.New("not found")
)

// PaginationParams defines the parameters for pagination
type PaginationParams struct {
	Page     int
	PageSize int
}

// UserRepository extends BaseRepository for User entities
type UserRepository interface {
	BaseRepository[entity.User]
}
