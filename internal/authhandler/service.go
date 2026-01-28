package authhandler

import (
	"context"
	"fmt"
	"time"

	"github.com/PersonaForge/backend/internal/auth"
	"github.com/PersonaForge/backend/internal/storage"
)

type authRepo interface {
	GetUserByGoogleID(googleID string) (*storage.User, error)
	CreateUser(googleID string, email string) (*storage.User, error)
}

type googleVerifier interface {
	VerifyIDToken(ctx context.Context, idToken string) (*auth.GoogleUserInfo, error)
}

type jwtIssuer interface {
	GenerateToken(userID int, email, sessionID string) (string, error)
	ExpiryMinutes() int
}

type chatSessionRepo interface {
	CreateSession(userID *int, sessionID string, isAnonymous bool, expiresAt time.Time) (int, error)
	MigrateSession(sessionID string, userID int) error
}

// Service handles authentication business logic
type Service struct {
	repo       authRepo
	googleAuth googleVerifier
	jwtService jwtIssuer
	chatRepo   chatSessionRepo
}

// NewService creates a new auth service
func NewService(repo *Repository, googleAuth *auth.GoogleAuthService, jwtService *auth.JWTService, chatRepo chatSessionRepo) *Service {
	return &Service{
		repo:       repo,
		googleAuth: googleAuth,
		jwtService: jwtService,
		chatRepo:   chatRepo,
	}
}

// GoogleLogin handles Google OAuth login
func (s *Service) GoogleLogin(ctx context.Context, idToken string, sessionID string) (*AuthResponse, error) {
	// Verify Google ID token
	userInfo, err := s.googleAuth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("invalid Google ID token: %w", err)
	}

	// Get or create user
	user, err := s.repo.GetUserByGoogleID(userInfo.GoogleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		// Create new user
		user, err = s.repo.CreateUser(userInfo.GoogleID, userInfo.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	// Generate new session ID if not provided
	var newSessionID string
	if sessionID != "" {
		// Migrate existing anonymous session
		err = s.chatRepo.MigrateSession(sessionID, user.ID)
		if err != nil {
			// Log error but continue (session might not exist)
			fmt.Printf("Warning: failed to migrate session: %v\n", err)
		}
		newSessionID = sessionID
	} else {
		newSessionID, err = auth.GenerateSessionID()
		if err != nil {
			return nil, fmt.Errorf("failed to generate session ID: %w", err)
		}
	}

	// Generate JWT
	token, err := s.jwtService.GenerateToken(user.ID, user.Email, newSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(s.jwtService.ExpiryMinutes()) * time.Minute)

	return &AuthResponse{
		Token:     token,
		UserID:    user.ID,
		Email:     user.Email,
		SessionID: newSessionID,
		ExpiresAt: expiresAt,
	}, nil
}

// CreateAnonymousSession creates a new anonymous session
func (s *Service) CreateAnonymousSession() (*AnonymousSessionResponse, error) {
	sessionID, err := auth.GenerateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	expiresAt := time.Now().Add(24 * time.Hour)

	// Create session in database
	_, err = s.chatRepo.CreateSession(nil, sessionID, true, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &AnonymousSessionResponse{
		SessionID: sessionID,
		ExpiresAt: expiresAt,
	}, nil
}

// ExpiryMinutes is a helper to expose JWT expiry
func (s *Service) ExpiryMinutes() int {
	return s.jwtService.ExpiryMinutes()
}
