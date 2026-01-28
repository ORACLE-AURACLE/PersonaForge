package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables
type Config struct {
	DatabaseURL      string
	JWTSecret        string
	GeminiAPIKey     string
	GoogleClientID   string
	Environment      string
	Port             string
	GeminiModel      string
	JWTExpiryMinutes int
}

// Load reads environment variables and returns a Config struct
// It fails fast if any required variable is missing
// It automatically loads .env.local if it exists (for local development)
// Supports both DATABASE_URL or individual DB_* variables
func Load() (*Config, error) {
	// Try to load .env.local for local development (ignore error if file doesn't exist)
	_ = godotenv.Load(".env.local")
	// Also try .env as fallback
	_ = godotenv.Load(".env")

	// Get database URL - either directly or build from individual variables
	databaseURL := getEnv("DATABASE_URL", "")
	if databaseURL == "" {
		// Build DATABASE_URL from individual DB_* variables
		dbHost := getEnv("DB_HOST", "localhost")
		dbPort := getEnv("DB_PORT", "5432")
		dbName := getEnv("DB_NAME", "personaforge")
		dbUser := getEnv("DB_USER", "postgres")
		dbPassword := getEnv("DB_PASSWORD", "")
		sslMode := getEnv("DB_SSLMODE", "disable")

		if dbPassword == "" {
			return nil, fmt.Errorf("either DATABASE_URL or DB_PASSWORD must be set")
		}

		databaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			dbUser, dbPassword, dbHost, dbPort, dbName, sslMode)
	}

	cfg := &Config{
		DatabaseURL:    databaseURL,
		JWTSecret:      getEnv("JWT_SECRET", ""),
		GeminiAPIKey:   getEnv("GEMINI_API_KEY", ""),
		GoogleClientID: getEnv("GOOGLE_CLIENT_ID", ""),
		Environment:    getEnv("ENV", "development"),
		Port:           getEnv("PORT", "8080"),
		GeminiModel:    getEnv("GEMINI_MODEL", "gemini-2.5-flash"),
	}

	// Parse JWT expiry
	expiryStr := getEnv("JWT_EXPIRY_MINUTES", "30")
	expiry, err := strconv.Atoi(expiryStr)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRY_MINUTES: %w", err)
	}
	cfg.JWTExpiryMinutes = expiry

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks that all required configuration values are present
func (c *Config) Validate() error {
	required := map[string]string{
		"DATABASE_URL":     c.DatabaseURL,
		"JWT_SECRET":       c.JWTSecret,
		"GEMINI_API_KEY":   c.GeminiAPIKey,
		"GOOGLE_CLIENT_ID": c.GoogleClientID,
	}

	for key, value := range required {
		if value == "" {
			return fmt.Errorf("required environment variable %s is not set", key)
		}
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
