package chat

import "time"

// SendMessageRequest is the request to send a chat message
type SendMessageRequest struct {
	PersonaID int    `json:"persona_id" binding:"required"`
	Message   string `json:"message" binding:"required"`
	SessionID string `json:"session_id,omitempty"` // Optional: provide to continue existing conversation, omit to start new
}

// ChatMessageResponse represents a single chat message
type ChatMessageResponse struct {
	ID        int    `json:"id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// SendMessageResponse is the response after sending a message
type SendMessageResponse struct {
	SessionID string              `json:"session_id"`
	Message   ChatMessageResponse `json:"message"`
}

// ConversationHistory represents the chat history
type ConversationHistory struct {
	SessionID string                `json:"session_id"`
	PersonaID int                   `json:"persona_id"`
	Messages  []ChatMessageResponse `json:"messages"`
}

// InsightRequest is the request for conversation insights
type InsightRequest struct {
	SessionID string `json:"session_id,omitempty"` // Optional: omit to use current session from auth
}

// InsightResponse is the response containing conversation insights
type InsightResponse struct {
	SessionID string `json:"session_id"`
	PersonaID int    `json:"persona_id"` // Profile (persona) in the conversation
	Analysis  string `json:"analysis"`
}

// MessageDTO is a data transfer object for messages
type MessageDTO struct {
	ID        int
	SessionID int
	PersonaID int
	Role      string
	Content   string
	CreatedAt time.Time
}
