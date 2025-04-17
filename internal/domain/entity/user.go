package entity

import (
	"time"
)

// User represents the core user entity
type User struct {
	ID        string    `json:"id" bigquery:"id"`
	Name      string    `json:"name" bigquery:"name"`
	Email     string    `json:"email" bigquery:"email"`
	CreatedAt time.Time `json:"created_at" bigquery:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bigquery:"updated_at"`
}
