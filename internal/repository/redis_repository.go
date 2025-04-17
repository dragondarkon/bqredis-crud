package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/dragondarkon/bqredis-crud/internal/domain/entity"
	"github.com/go-redis/redis/v8"
)

const (
	// Cache key prefixes
	userKeyPrefix     = "users:"
	userListKeyPrefix = "users:list:"
	pageKeyFormat     = "page_%d:size_%d"
	defaultTimeout    = 3 * time.Second
)

// RedisRepository implements a caching layer over another UserRepository
type RedisRepository struct {
	BaseRepositoryImpl[entity.User]
	client     *redis.Client
	repository UserRepository
	ttl        time.Duration
}

// NewRedisRepository creates a new Redis repository
func NewRedisRepository(client *redis.Client, repository UserRepository, ttl time.Duration) *RedisRepository {
	return &RedisRepository{
		client:     client,
		repository: repository,
		ttl:        ttl,
	}
}

// executeWithTimeout executes a Redis operation with a timeout
func (r *RedisRepository) executeWithTimeout(ctx context.Context, operation func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- operation(ctx)
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return fmt.Errorf("redis operation timed out: %w", ctx.Err())
	}
}

// cacheGet retrieves a value from Redis and unmarshals it
func (r *RedisRepository) cacheGet(ctx context.Context, key string, result interface{}) error {
	var data string
	err := r.executeWithTimeout(ctx, func(ctx context.Context) error {
		var err error
		data, err = r.client.Get(ctx, key).Result()
		if err == redis.Nil {
			return fmt.Errorf("cache miss for key %s", key)
		}
		return err
	})
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(data), result)
}

// cacheSet stores a value in Redis with the configured TTL
func (r *RedisRepository) cacheSet(ctx context.Context, key string, value interface{}) error {
	return r.executeWithTimeout(ctx, func(ctx context.Context) error {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal data: %w", err)
		}
		return r.client.Set(ctx, key, data, r.ttl).Err()
	})
}

// generateKey creates cache keys for different types of data
func (r *RedisRepository) generateKey(id string) string {
	return userKeyPrefix + id
}

func (r *RedisRepository) generateListKey(params PaginationParams) string {
	return fmt.Sprintf("%s%s", userListKeyPrefix, fmt.Sprintf(pageKeyFormat, params.Page, params.PageSize))
}

// invalidateCache removes user-related cache entries
func (r *RedisRepository) invalidateCache(ctx context.Context, id string) error {
	return r.executeWithTimeout(ctx, func(ctx context.Context) error {
		pipe := r.client.Pipeline()
		pipe.Del(ctx, r.generateKey(id))
		pipe.Del(ctx, userListKeyPrefix+"*")
		_, err := pipe.Exec(ctx)
		return err
	})
}

// GetAll retrieves all users with pagination, using cache if possible
func (r *RedisRepository) GetAll(ctx context.Context, params PaginationParams) ([]entity.User, error) {
	r.ValidatePagination(&params)
	cacheKey := r.generateListKey(params)

	var users []entity.User
	err := r.cacheGet(ctx, cacheKey, &users)
	if err == nil {
		return users, nil
	}

	// Cache miss, get from underlying repository
	users, err = r.repository.GetAll(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get users from repository: %w", err)
	}

	// Update cache in background
	go func() {
		if err := r.cacheSet(context.Background(), cacheKey, users); err != nil {
			log.Printf("Failed to cache users list: %v", err)
		}
	}()

	return users, nil
}

// GetByID retrieves a user by ID, using cache if possible
func (r *RedisRepository) GetByID(ctx context.Context, id string) (entity.User, error) {
	if err := r.ValidateID(id); err != nil {
		return entity.User{}, err
	}

	cacheKey := r.generateKey(id)
	var user entity.User
	err := r.cacheGet(ctx, cacheKey, &user)
	if err == nil {
		return user, nil
	}

	// Cache miss, get from underlying repository
	user, err = r.repository.GetByID(ctx, id)
	if err != nil {
		return entity.User{}, fmt.Errorf("failed to get user from repository: %w", err)
	}

	// Update cache in background
	go func() {
		if err := r.cacheSet(context.Background(), cacheKey, user); err != nil {
			log.Printf("Failed to cache user: %v", err)
		}
	}()

	return user, nil
}

// Create creates a user and updates cache
func (r *RedisRepository) Create(ctx context.Context, user entity.User) error {
	if err := r.repository.Create(ctx, user); err != nil {
		return fmt.Errorf("failed to create user in repository: %w", err)
	}

	if err := r.invalidateCache(ctx, user.ID); err != nil {
		log.Printf("Failed to invalidate cache after create: %v", err)
	}

	return nil
}

// Update updates a user and updates cache
func (r *RedisRepository) Update(ctx context.Context, user entity.User) error {
	if err := r.ValidateID(user.ID); err != nil {
		return err
	}

	if err := r.repository.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user in repository: %w", err)
	}

	if err := r.invalidateCache(ctx, user.ID); err != nil {
		log.Printf("Failed to invalidate cache after update: %v", err)
	}

	return nil
}

// Delete removes a user and updates cache
func (r *RedisRepository) Delete(ctx context.Context, id string) error {
	if err := r.ValidateID(id); err != nil {
		return err
	}

	if err := r.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user from repository: %w", err)
	}

	if err := r.invalidateCache(ctx, id); err != nil {
		log.Printf("Failed to invalidate cache after delete: %v", err)
	}

	return nil
}
