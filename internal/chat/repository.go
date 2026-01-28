package chat

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/PersonaForge/backend/internal/storage"
)

// Repository handles chat data access
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new chat repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateSession creates a new chat session
func (r *Repository) CreateSession(userID *int, sessionID string, isAnonymous bool, expiresAt time.Time) (int, error) {
	var id int
	query := `
		INSERT INTO sessions (user_id, session_id, is_anonymous, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err := r.db.QueryRow(query, userID, sessionID, isAnonymous, time.Now(), expiresAt).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create session: %w", err)
	}

	return id, nil
}

// GetSessionByID retrieves a session by session ID string
func (r *Repository) GetSessionByID(sessionID string) (*storage.Session, error) {
	var session storage.Session
	query := `
		SELECT id, user_id, session_id, is_anonymous, created_at, expires_at
		FROM sessions
		WHERE session_id = $1
	`

	err := r.db.QueryRow(query, sessionID).Scan(
		&session.ID,
		&session.UserID,
		&session.SessionID,
		&session.IsAnonymous,
		&session.CreatedAt,
		&session.ExpiresAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// MigrateSession attaches an anonymous session to a user
func (r *Repository) MigrateSession(sessionID string, userID int) error {
	query := `
		UPDATE sessions
		SET user_id = $1, is_anonymous = false
		WHERE session_id = $2 AND is_anonymous = true
	`

	_, err := r.db.Exec(query, userID, sessionID)
	if err != nil {
		return fmt.Errorf("failed to migrate session: %w", err)
	}

	return nil
}

// SaveMessage saves a chat message
func (r *Repository) SaveMessage(sessionDBID int, personaID int, role string, content string) (*MessageDTO, error) {
	var msg MessageDTO
	query := `
		INSERT INTO messages (session_id, persona_id, role, content, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, session_id, persona_id, role, content, created_at
	`

	err := r.db.QueryRow(query, sessionDBID, personaID, role, content, time.Now()).Scan(
		&msg.ID,
		&msg.SessionID,
		&msg.PersonaID,
		&msg.Role,
		&msg.Content,
		&msg.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	return &msg, nil
}

// GetConversationHistory retrieves all messages for a session
func (r *Repository) GetConversationHistory(sessionDBID int, personaID int) ([]MessageDTO, error) {
	query := `
		SELECT id, session_id, persona_id, role, content, created_at
		FROM messages
		WHERE session_id = $1 AND persona_id = $2
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, sessionDBID, personaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation history: %w", err)
	}
	defer rows.Close()

	var messages []MessageDTO
	for rows.Next() {
		var msg MessageDTO
		err := rows.Scan(
			&msg.ID,
			&msg.SessionID,
			&msg.PersonaID,
			&msg.Role,
			&msg.Content,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

// GetAllMessagesForSession retrieves all messages for a session across all personas.
func (r *Repository) GetAllMessagesForSession(sessionDBID int) ([]MessageDTO, error) {
	query := `
		SELECT id, session_id, persona_id, role, content, created_at
		FROM messages
		WHERE session_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, sessionDBID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	var messages []MessageDTO
	for rows.Next() {
		var msg MessageDTO
		err := rows.Scan(&msg.ID, &msg.SessionID, &msg.PersonaID, &msg.Role, &msg.Content, &msg.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

// SaveTokenUsage records token usage for a session
func (r *Repository) SaveTokenUsage(sessionDBID int, promptTokens int, completionTokens int, totalTokens int) error {
	query := `
		INSERT INTO token_usage (session_id, prompt_tokens, completion_tokens, total_tokens, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(query, sessionDBID, promptTokens, completionTokens, totalTokens, time.Now())
	if err != nil {
		return fmt.Errorf("failed to save token usage: %w", err)
	}

	return nil
}
