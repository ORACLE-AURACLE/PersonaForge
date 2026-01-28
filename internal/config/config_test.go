package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Set required env vars
	os.Setenv("USE_MONGO", "false") // Explicitly set to false to test PostgreSQL path
	os.Setenv("DATABASE_URL", "postgresql://localhost/test")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("GEMINI_API_KEY", "test-key")
	os.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
	defer func() {
		os.Unsetenv("USE_MONGO")
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
		UseMongo:    false,
		DatabaseURL: "postgresql://localhost/test",
		// Missing other required fields
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected validation error for missing required fields")
	}
}

func TestLoad_MongoDB(t *testing.T) {
	// Set required env vars for MongoDB
	os.Setenv("USE_MONGO", "true")
	os.Setenv("MONGODB_URI", "mongodb://localhost:27017/test")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("GEMINI_API_KEY", "test-key")
	os.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
	defer func() {
		os.Unsetenv("USE_MONGO")
		os.Unsetenv("MONGODB_URI")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("GEMINI_API_KEY")
		os.Unsetenv("GOOGLE_CLIENT_ID")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !cfg.UseMongo {
		t.Error("expected UseMongo to be true")
	}

	if cfg.MongoURI != "mongodb://localhost:27017/test" {
		t.Errorf("expected MongoURI to be 'mongodb://localhost:27017/test', got %s", cfg.MongoURI)
	}
}
