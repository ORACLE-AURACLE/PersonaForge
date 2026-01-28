package authhandler

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/PersonaForge/backend/internal/storage"
)

// Repository handles authentication data access
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new auth repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetUserByGoogleID retrieves a user by Google ID
func (r *Repository) GetUserByGoogleID(googleID string) (*storage.User, error) {
	var user storage.User
	query := `
		SELECT id, google_id, email, created_at, updated_at
		FROM users
		WHERE google_id = $1
	`

	err := r.db.QueryRow(query, googleID).Scan(
		&user.ID,
		&user.GoogleID,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// CreateUser creates a new user
func (r *Repository) CreateUser(googleID string, email string) (*storage.User, error) {
	var user storage.User
	query := `
		INSERT INTO users (google_id, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, google_id, email, created_at, updated_at
	`

	now := time.Now()
	err := r.db.QueryRow(query, googleID, email, now, now).Scan(
		&user.ID,
		&user.GoogleID,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}
