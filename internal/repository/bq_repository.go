package repository

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/dragondarkon/bqredis-crud/internal/domain/entity"
	"google.golang.org/api/iterator"
)

// BigQueryRepository implements UserRepository using BigQuery
type BigQueryRepository struct {
	BaseRepositoryImpl[entity.User]
	client    *bigquery.Client
	projectID string
	dataset   string
	table     string
}

// NewBigQueryRepository creates a new BigQuery repository
func NewBigQueryRepository(client *bigquery.Client, projectID, dataset, table string) *BigQueryRepository {
	return &BigQueryRepository{
		client:    client,
		projectID: projectID,
		dataset:   dataset,
		table:     table,
	}
}

// GetAll retrieves all users from BigQuery with pagination
func (r *BigQueryRepository) GetAll(ctx context.Context, params PaginationParams) ([]entity.User, error) {
	r.ValidatePagination(&params)
	offset := r.CalculateOffset(params)

	query := r.client.Query(`
		SELECT id, name, email, created_at, updated_at
		FROM @dataset.@table
		ORDER BY created_at DESC
		LIMIT @pageSize
		OFFSET @offset
	`)
	query.Parameters = []bigquery.QueryParameter{
		{Name: "dataset", Value: r.dataset},
		{Name: "table", Value: r.table},
		{Name: "pageSize", Value: params.PageSize},
		{Name: "offset", Value: offset},
	}

	return r.executeQuery(ctx, query)
}

// GetByID retrieves a user by ID from BigQuery
func (r *BigQueryRepository) GetByID(ctx context.Context, id string) (entity.User, error) {
	if err := r.ValidateID(id); err != nil {
		return entity.User{}, err
	}

	query := r.client.Query(`
		SELECT id, name, email, created_at, updated_at
		FROM @dataset.@table
		WHERE id = @id
	`)
	query.Parameters = []bigquery.QueryParameter{
		{Name: "dataset", Value: r.dataset},
		{Name: "table", Value: r.table},
		{Name: "id", Value: id},
	}

	users, err := r.executeQuery(ctx, query)
	if err != nil {
		return entity.User{}, err
	}
	if len(users) == 0 {
		return entity.User{}, fmt.Errorf("user %s: %w", id, ErrNotFound)
	}
	return users[0], nil
}

// executeQuery is a helper method to execute BigQuery queries and return users
func (r *BigQueryRepository) executeQuery(ctx context.Context, query *bigquery.Query) ([]entity.User, error) {
	it, err := query.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	var users []entity.User
	for {
		var user entity.User
		err := it.Next(&user)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// Create inserts a new user into BigQuery
func (r *BigQueryRepository) Create(ctx context.Context, user entity.User) error {
	inserter := r.client.Dataset(r.dataset).Table(r.table).Inserter()
	if err := inserter.Put(ctx, user); err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}
	return nil
}

// Update updates an existing user in BigQuery
func (r *BigQueryRepository) Update(ctx context.Context, user entity.User) error {
	if err := r.ValidateID(user.ID); err != nil {
		return err
	}

	// First check if user exists
	_, err := r.GetByID(ctx, user.ID)
	if err != nil {
		return err
	}

	query := r.client.Query(`
		UPDATE @dataset.@table
		SET name = @name, 
			email = @email, 
			updated_at = @updatedAt
		WHERE id = @id
	`)
	query.Parameters = []bigquery.QueryParameter{
		{Name: "dataset", Value: r.dataset},
		{Name: "table", Value: r.table},
		{Name: "name", Value: user.Name},
		{Name: "email", Value: user.Email},
		{Name: "updatedAt", Value: user.UpdatedAt},
		{Name: "id", Value: user.ID},
	}

	return r.executeUpdateQuery(ctx, query)
}

// Delete removes a user from BigQuery
func (r *BigQueryRepository) Delete(ctx context.Context, id string) error {
	if err := r.ValidateID(id); err != nil {
		return err
	}

	// First check if user exists
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	query := r.client.Query(`
		DELETE FROM @dataset.@table
		WHERE id = @id
	`)
	query.Parameters = []bigquery.QueryParameter{
		{Name: "dataset", Value: r.dataset},
		{Name: "table", Value: r.table},
		{Name: "id", Value: id},
	}

	return r.executeUpdateQuery(ctx, query)
}

// executeUpdateQuery is a helper method to execute update/delete queries
func (r *BigQueryRepository) executeUpdateQuery(ctx context.Context, query *bigquery.Query) error {
	job, err := query.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	_, err = job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to complete job: %w", err)
	}

	return nil
}
