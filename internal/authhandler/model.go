package authhandler

import "time"

// GoogleLoginRequest is the request to login with Google
type GoogleLoginRequest struct {
	IDToken   string `json:"id_token" binding:"required"`
	SessionID string `json:"session_id,omitempty"` // Optional: for migrating anonymous sessions
}

// AnonymousSessionRequest is the request to create an anonymous session
type AnonymousSessionRequest struct {
	// No fields needed, session ID is generated
}

// AuthResponse is the response after successful authentication
type AuthResponse struct {
	Token     string    `json:"token"`
	UserID    int       `json:"user_id"`
	Email     string    `json:"email"`
	SessionID string    `json:"session_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// AnonymousSessionResponse is the response after creating an anonymous session
type AnonymousSessionResponse struct {
	SessionID string    `json:"session_id"`
	ExpiresAt time.Time `json:"expires_at"`
}
