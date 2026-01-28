package storage

import "time"

// User represents a registered user
type User struct {
	ID        int       `json:"id" db:"id" bson:"_id,omitempty"`
	GoogleID  string    `json:"google_id" db:"google_id" bson:"google_id"`
	Email     string    `json:"email" db:"email" bson:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" bson:"updated_at"`
}

// Persona represents an AI persona blueprint
type Persona struct {
	ID        int       `json:"id" db:"id" bson:"_id,omitempty"`
	UserID    *int      `json:"user_id,omitempty" db:"user_id" bson:"user_id,omitempty"`
	SessionID *string   `json:"session_id,omitempty" db:"session_id" bson:"session_id,omitempty"`
	Name      string    `json:"name" db:"name" bson:"name"`
	Blueprint string    `json:"blueprint" db:"blueprint" bson:"blueprint"` // JSON string
	IsDefault bool      `json:"is_default" db:"is_default" bson:"is_default"`
	CreatedAt time.Time `json:"created_at" db:"created_at" bson:"created_at"`
}

// Session represents a user session (anonymous or authenticated)
type Session struct {
	ID          int       `json:"id" db:"id" bson:"_id,omitempty"`
	UserID      *int      `json:"user_id,omitempty" db:"user_id" bson:"user_id,omitempty"`
	SessionID   string    `json:"session_id" db:"session_id" bson:"session_id"`
	IsAnonymous bool      `json:"is_anonymous" db:"is_anonymous" bson:"is_anonymous"`
	CreatedAt   time.Time `json:"created_at" db:"created_at" bson:"created_at"`
	ExpiresAt   time.Time `json:"expires_at" db:"expires_at" bson:"expires_at"`
}

// Message represents a chat message
type Message struct {
	ID        int       `json:"id" db:"id" bson:"_id,omitempty"`
	SessionID int       `json:"session_id" db:"session_id" bson:"session_id"`
	PersonaID int       `json:"persona_id" db:"persona_id" bson:"persona_id"`
	Role      string    `json:"role" db:"role" bson:"role"` // user, assistant, system
	Content   string    `json:"content" db:"content" bson:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at" bson:"created_at"`
}

// TokenUsage tracks Gemini API token consumption
type TokenUsage struct {
	ID               int       `json:"id" db:"id" bson:"_id,omitempty"`
	SessionID        int       `json:"session_id" db:"session_id" bson:"session_id"`
	PromptTokens     int       `json:"prompt_tokens" db:"prompt_tokens" bson:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens" db:"completion_tokens" bson:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens" db:"total_tokens" bson:"total_tokens"`
	CreatedAt        time.Time `json:"created_at" db:"created_at" bson:"created_at"`
}
