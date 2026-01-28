package storage

import "time"

// User represents a registered user
type User struct {
	ID        int       `json:"id" db:"id"`
	GoogleID  string    `json:"google_id" db:"google_id"`
	Email     string    `json:"email" db:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Persona represents an AI persona blueprint
type Persona struct {
	ID        int       `json:"id" db:"id"`
	UserID    *int      `json:"user_id,omitempty" db:"user_id"`
	SessionID *string   `json:"session_id,omitempty" db:"session_id"`
	Name      string    `json:"name" db:"name"`
	Blueprint string    `json:"blueprint" db:"blueprint"` // JSON string
	IsDefault bool      `json:"is_default" db:"is_default"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Session represents a user session (anonymous or authenticated)
type Session struct {
	ID          int       `json:"id" db:"id"`
	UserID      *int      `json:"user_id,omitempty" db:"user_id"`
	SessionID   string    `json:"session_id" db:"session_id"`
	IsAnonymous bool      `json:"is_anonymous" db:"is_anonymous"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	ExpiresAt   time.Time `json:"expires_at" db:"expires_at"`
}

// Message represents a chat message
type Message struct {
	ID        int       `json:"id" db:"id"`
	SessionID int       `json:"session_id" db:"session_id"`
	PersonaID int       `json:"persona_id" db:"persona_id"`
	Role      string    `json:"role" db:"role"` // user, assistant, system
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// TokenUsage tracks Gemini API token consumption
type TokenUsage struct {
	ID               int       `json:"id" db:"id"`
	SessionID        int       `json:"session_id" db:"session_id"`
	PromptTokens     int       `json:"prompt_tokens" db:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens" db:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens" db:"total_tokens"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}
