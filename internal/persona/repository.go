package persona

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/PersonaForge/backend/internal/storage"
)

// Repository handles persona data access for PostgreSQL
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new persona repository for PostgreSQL
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// MongoRepository handles persona data access for MongoDB
type MongoRepository struct {
	db *storage.MongoDatabase
}

// NewMongoRepository creates a new persona repository for MongoDB
func NewMongoRepository(db *storage.MongoDatabase) *MongoRepository {
	return &MongoRepository{db: db}
}

// SessionActive checks whether a session exists and is not expired.
func (r *Repository) SessionActive(sessionID string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM sessions
			WHERE session_id = $1 AND expires_at > NOW()
		)
	`
	if err := r.db.QueryRow(query, sessionID).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check session: %w", err)
	}
	return exists, nil
}

// CreatePersona creates a new custom persona
func (r *Repository) CreatePersona(userID *int, sessionID *string, name string, blueprint string) (*storage.Persona, error) {
	var persona storage.Persona

	query := `
		INSERT INTO personas (user_id, session_id, name, blueprint, is_default, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, session_id, name, blueprint, is_default, created_at
	`

	err := r.db.QueryRow(query, userID, sessionID, name, blueprint, false, time.Now()).Scan(
		&persona.ID,
		&persona.UserID,
		&persona.SessionID,
		&persona.Name,
		&persona.Blueprint,
		&persona.IsDefault,
		&persona.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create persona: %w", err)
	}

	return &persona, nil
}

// GetPersonaByID retrieves a persona by ID
func (r *Repository) GetPersonaByID(id int) (*storage.Persona, error) {
	var persona storage.Persona

	query := `
		SELECT id, user_id, session_id, name, blueprint, is_default, created_at
		FROM personas
		WHERE id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&persona.ID,
		&persona.UserID,
		&persona.SessionID,
		&persona.Name,
		&persona.Blueprint,
		&persona.IsDefault,
		&persona.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("persona not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get persona: %w", err)
	}

	return &persona, nil
}

// ListDefaultPersonas retrieves all default personas.
func (r *Repository) ListDefaultPersonas() ([]storage.Persona, error) {
	query := `
		SELECT id, user_id, session_id, name, blueprint, is_default, created_at
		FROM personas
		WHERE is_default = true
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list personas: %w", err)
	}
	defer rows.Close()

	var personas []storage.Persona
	for rows.Next() {
		var persona storage.Persona
		err := rows.Scan(
			&persona.ID,
			&persona.UserID,
			&persona.SessionID,
			&persona.Name,
			&persona.Blueprint,
			&persona.IsDefault,
			&persona.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan persona: %w", err)
		}
		personas = append(personas, persona)
	}

	return personas, rows.Err()
}

// ListPersonasForUser retrieves all personas for a user (including defaults)
func (r *Repository) ListPersonasForUser(userID int) ([]storage.Persona, error) {
	query := `
		SELECT id, user_id, session_id, name, blueprint, is_default, created_at
		FROM personas
		WHERE is_default = true OR user_id = $1
		ORDER BY is_default DESC, created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list personas: %w", err)
	}
	defer rows.Close()

	var personas []storage.Persona
	for rows.Next() {
		var persona storage.Persona
		err := rows.Scan(
			&persona.ID,
			&persona.UserID,
			&persona.SessionID,
			&persona.Name,
			&persona.Blueprint,
			&persona.IsDefault,
			&persona.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan persona: %w", err)
		}
		personas = append(personas, persona)
	}

	return personas, rows.Err()
}

// ListPersonasForSession retrieves all personas for a guest session (including defaults)
func (r *Repository) ListPersonasForSession(sessionID string) ([]storage.Persona, error) {
	query := `
		SELECT id, user_id, session_id, name, blueprint, is_default, created_at
		FROM personas
		WHERE is_default = true OR session_id = $1
		ORDER BY is_default DESC, created_at DESC
	`

	rows, err := r.db.Query(query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list personas: %w", err)
	}
	defer rows.Close()

	var personas []storage.Persona
	for rows.Next() {
		var persona storage.Persona
		err := rows.Scan(
			&persona.ID,
			&persona.UserID,
			&persona.SessionID,
			&persona.Name,
			&persona.Blueprint,
			&persona.IsDefault,
			&persona.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan persona: %w", err)
		}
		personas = append(personas, persona)
	}

	return personas, rows.Err()
}

// ListCustomPersonasForSession returns only custom personas created by the given session.
func (r *Repository) ListCustomPersonasForSession(sessionID string) ([]storage.Persona, error) {
	query := `
		SELECT id, user_id, session_id, name, blueprint, is_default, created_at
		FROM personas
		WHERE session_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list personas for session: %w", err)
	}
	defer rows.Close()

	var personas []storage.Persona
	for rows.Next() {
		var persona storage.Persona
		err := rows.Scan(
			&persona.ID,
			&persona.UserID,
			&persona.SessionID,
			&persona.Name,
			&persona.Blueprint,
			&persona.IsDefault,
			&persona.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan persona: %w", err)
		}
		personas = append(personas, persona)
	}
	return personas, rows.Err()
}

// CountCustomPersonasForUser counts non-default personas for a user
func (r *Repository) CountCustomPersonasForUser(userID int) (int, error) {
	var count int

	query := `
		SELECT COUNT(*)
		FROM personas
		WHERE user_id = $1 AND is_default = false
	`

	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count personas: %w", err)
	}

	return count, nil
}

// CountCustomPersonasForSession counts non-default personas for a guest session
func (r *Repository) CountCustomPersonasForSession(sessionID string) (int, error) {
	var count int

	query := `
		SELECT COUNT(*)
		FROM personas
		WHERE session_id = $1 AND is_default = false
	`

	err := r.db.QueryRow(query, sessionID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count personas: %w", err)
	}

	return count, nil
}

// DeletePersona deletes a custom persona
func (r *Repository) DeletePersona(id int, userID int) error {
	query := `
		DELETE FROM personas
		WHERE id = $1 AND user_id = $2 AND is_default = false
	`

	result, err := r.db.Exec(query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete persona: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("persona not found or cannot be deleted")
	}

	return nil
}

// InitializeDefaultPersonas creates the 4 default personas if they don't exist
func (r *Repository) InitializeDefaultPersonas() error {
	// Check if defaults already exist
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM personas WHERE is_default = true").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check default personas: %w", err)
	}

	if count > 0 {
		return nil // Already initialized
	}

	// Create default personas
	defaults := DefaultPersonas()
	for _, blueprint := range defaults {
		blueprintJSON, err := MarshalBlueprint(blueprint)
		if err != nil {
			return fmt.Errorf("failed to marshal blueprint: %w", err)
		}

		query := `
			INSERT INTO personas (user_id, name, blueprint, is_default, created_at)
			VALUES (NULL, $1, $2, true, $3)
		`

		_, err = r.db.Exec(query, blueprint.Name, blueprintJSON, time.Now())
		if err != nil {
			return fmt.Errorf("failed to create default persona: %w", err)
		}
	}

	return nil
}
