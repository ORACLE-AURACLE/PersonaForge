package auth

import (
	"context"
	"fmt"

	"google.golang.org/api/idtoken"
)

// GoogleAuthService handles Google OAuth verification
type GoogleAuthService struct {
	clientID string
}

// NewGoogleAuthService creates a new Google auth service
func NewGoogleAuthService(clientID string) *GoogleAuthService {
	return &GoogleAuthService{
		clientID: clientID,
	}
}

// GoogleUserInfo contains user information from Google
type GoogleUserInfo struct {
	GoogleID string
	Email    string
	Name     string
}

// VerifyIDToken verifies a Google ID token and extracts user info
func (g *GoogleAuthService) VerifyIDToken(ctx context.Context, idToken string) (*GoogleUserInfo, error) {
	payload, err := idtoken.Validate(ctx, idToken, g.clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate ID token: %w", err)
	}

	// Extract user information
	email, ok := payload.Claims["email"].(string)
	if !ok {
		return nil, fmt.Errorf("email not found in token")
	}

	sub, ok := payload.Claims["sub"].(string)
	if !ok {
		return nil, fmt.Errorf("subject (user ID) not found in token")
	}

	name, _ := payload.Claims["name"].(string)

	return &GoogleUserInfo{
		GoogleID: sub,
		Email:    email,
		Name:     name,
	}, nil
}
