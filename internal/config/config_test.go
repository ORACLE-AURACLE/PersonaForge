package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Set required env vars
	os.Setenv("DATABASE_URL", "postgresql://localhost/test")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("GEMINI_API_KEY", "test-key")
	os.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("GEMINI_API_KEY")
		os.Unsetenv("GOOGLE_CLIENT_ID")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.DatabaseURL != "postgresql://localhost/test" {
		t.Errorf("expected DatabaseURL to be 'postgresql://localhost/test', got %s", cfg.DatabaseURL)
	}

	if cfg.GeminiModel != "gemini-2.5-flash" {
		t.Errorf("expected default GeminiModel to be 'gemini-2.5-flash', got %s", cfg.GeminiModel)
	}
}

func TestValidate_MissingRequired(t *testing.T) {
	cfg := &Config{
		DatabaseURL: "postgresql://localhost/test",
		// Missing other required fields
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected validation error for missing required fields")
	}
}
