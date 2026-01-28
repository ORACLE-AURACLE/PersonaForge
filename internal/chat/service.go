package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/PersonaForge/backend/internal/auth"
	"github.com/PersonaForge/backend/internal/gemini"
	"github.com/PersonaForge/backend/internal/persona"
	"github.com/PersonaForge/backend/internal/storage"
)

type chatRepo interface {
	CreateSession(userID *int, sessionID string, isAnonymous bool, expiresAt time.Time) (int, error)
	GetSessionByID(sessionID string) (*storage.Session, error)
	GetConversationHistory(sessionDBID int, personaID int) ([]MessageDTO, error)
	GetAllMessagesForSession(sessionDBID int) ([]MessageDTO, error)
	SaveMessage(sessionDBID int, personaID int, role string, content string) (*MessageDTO, error)
	SaveTokenUsage(sessionDBID int, promptTokens int, completionTokens int, totalTokens int) error
	MigrateSession(sessionID string, userID int) error
}

type personaBlueprintProvider interface {
	GetPersonaBlueprint(id int) (*persona.PersonaBlueprint, error)
}

type geminiClient interface {
	GenerateContent(ctx context.Context, req gemini.GenerateContentRequest) (*gemini.GenerateContentResponse, error)
	GenerateInsight(ctx context.Context, messages []gemini.Message) (*gemini.InsightResponse, error)
}

// Service handles chat business logic
type Service struct {
	repo           chatRepo
	geminiClient   geminiClient
	personaService personaBlueprintProvider
}

// NewService creates a new chat service
func NewService(repo *Repository, _ *persona.Repository, geminiClient *gemini.Client, personaService *persona.Service) *Service {
	return &Service{
		repo:           repo,
		geminiClient:   geminiClient,
		personaService: personaService,
	}
}

// SendMessage handles the complete chat flow
func (s *Service) SendMessage(ctx context.Context, userID *int, isAuthenticated bool, req SendMessageRequest) (*SendMessageResponse, error) {
	// 1. Identify or create session
	var sessionDBID int
	var sessionIDStr string
	var err error

	if req.SessionID != "" {
		// Use existing session
		session, err := s.repo.GetSessionByID(req.SessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to get session: %w", err)
		}
		if session == nil {
			return nil, fmt.Errorf("session not found")
		}
		// Ownership checks
		if isAuthenticated {
			if userID == nil || session.UserID == nil || *session.UserID != *userID {
				return nil, fmt.Errorf("session not found")
			}
			if session.ExpiresAt.Before(time.Now()) {
				return nil, fmt.Errorf("session not found")
			}
		} else {
			// Guest: only allow anonymous sessions
			if !session.IsAnonymous {
				return nil, fmt.Errorf("session not found")
			}
			if session.ExpiresAt.Before(time.Now()) {
				return nil, fmt.Errorf("session not found")
			}
		}
		sessionDBID = session.ID
		sessionIDStr = session.SessionID
	} else {
		// Create new session
		sessionIDStr, err = auth.GenerateSessionID()
		if err != nil {
			return nil, fmt.Errorf("failed to generate session ID: %w", err)
		}

		expiresAt := time.Now().Add(24 * time.Hour)
		sessionDBID, err = s.repo.CreateSession(userID, sessionIDStr, !isAuthenticated, expiresAt)
		if err != nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}
	}

	// 2. Load persona blueprint
	personaBlueprint, err := s.personaService.GetPersonaBlueprint(req.PersonaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get persona: %w", err)
	}

	// 3. Load conversation history
	history, err := s.repo.GetConversationHistory(sessionDBID, req.PersonaID)
	if err != nil {
		return nil, fmt.Errorf("failed to load conversation history: %w", err)
	}

	// 4. Save user message
	_, err = s.repo.SaveMessage(sessionDBID, req.PersonaID, "user", req.Message)
	if err != nil {
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}

	// 5. Build Gemini request
	var messages []gemini.Message
	for _, msg := range history {
		messages = append(messages, gemini.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	geminiReq := gemini.GenerateContentRequest{
		SystemPrompt: personaBlueprint.ToSystemPrompt(),
		Messages:     messages,
		UserMessage:  req.Message,
	}

	// 6. Call Gemini
	response, err := s.geminiClient.GenerateContent(ctx, geminiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	// 7. Save assistant response
	assistantMsg, err := s.repo.SaveMessage(sessionDBID, req.PersonaID, "assistant", response.Text)
	if err != nil {
		return nil, fmt.Errorf("failed to save assistant message: %w", err)
	}

	// 8. Track token usage
	err = s.repo.SaveTokenUsage(sessionDBID, response.PromptTokens, response.OutputTokens, response.TotalTokens)
	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to save token usage: %v\n", err)
	}

	return &SendMessageResponse{
		SessionID: sessionIDStr,
		Message: ChatMessageResponse{
			ID:        assistantMsg.ID,
			Role:      assistantMsg.Role,
			Content:   assistantMsg.Content,
			CreatedAt: assistantMsg.CreatedAt.Format("2006-01-02T15:04:05Z"),
		},
	}, nil
}

// GetConversationHistory retrieves the conversation history for an authenticated user and session
func (s *Service) GetConversationHistory(ctx context.Context, userID int, sessionID string, personaID int) (*ConversationHistory, error) {
	session, err := s.repo.GetSessionByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	if session == nil {
		return nil, fmt.Errorf("session not found")
	}
	if session.UserID == nil || *session.UserID != userID {
		return nil, fmt.Errorf("session not found")
	}
	if session.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("session not found")
	}

	messages, err := s.repo.GetConversationHistory(session.ID, personaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation history: %w", err)
	}

	var chatMessages []ChatMessageResponse
	for _, msg := range messages {
		chatMessages = append(chatMessages, ChatMessageResponse{
			ID:        msg.ID,
			Role:      msg.Role,
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &ConversationHistory{
		SessionID: sessionID,
		PersonaID: personaID,
		Messages:  chatMessages,
	}, nil
}

// GenerateInsight generates insights for a conversation
func (s *Service) GenerateInsight(ctx context.Context, userID int, sessionID string) (*InsightResponse, error) {
	session, err := s.repo.GetSessionByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	if session == nil {
		return nil, fmt.Errorf("session not found")
	}
	if session.UserID == nil || *session.UserID != userID {
		return nil, fmt.Errorf("session not found")
	}
	if session.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("session not found")
	}

	all, err := s.repo.GetAllMessagesForSession(session.ID)
	if err != nil {
		return nil, err
	}

	var messages []gemini.Message
	for _, msg := range all {
		messages = append(messages, gemini.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages found in session")
	}

	// Generate insight
	insight, err := s.geminiClient.GenerateInsight(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to generate insight: %w", err)
	}

	return &InsightResponse{
		SessionID: sessionID,
		Analysis:  insight.Analysis,
	}, nil
}

// MigrateSession migrates an anonymous session to an authenticated user
func (s *Service) MigrateSession(sessionID string, userID int) error {
	return s.repo.MigrateSession(sessionID, userID)
}
