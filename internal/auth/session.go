package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateSessionID creates a random session ID
func GenerateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate session ID: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
