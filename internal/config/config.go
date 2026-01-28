package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables
type Config struct {
	UseMongo         bool
	DatabaseURL      string
	MongoURI         string
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

	// Check if MongoDB should be used
	useMongoStr := getEnv("USE_MONGO", "true")
	useMongo, err := strconv.ParseBool(useMongoStr)
	if err != nil {
		useMongo = true // Default to MongoDB if parsing fails
	}

	var databaseURL string
	var mongoURI string

	if useMongo {
		mongoURI = getEnv("MONGODB_URI", "")
		if mongoURI == "" {
			return nil, fmt.Errorf("MONGODB_URI is required when USE_MONGO=true")
		}
	} else {
		// Get database URL - either directly or build from individual variables
		databaseURL = getEnv("DATABASE_URL", "")
		if databaseURL == "" {
			// Build DATABASE_URL from individual DB_* variables
			dbHost := getEnv("DB_HOST", "localhost")
			dbPort := getEnv("DB_PORT", "5432")
			dbName := getEnv("DB_NAME", "personaforge")
			dbUser := getEnv("DB_USER", "postgres")
			dbPassword := getEnv("DB_PASSWORD", "")
			sslMode := getEnv("DB_SSLMODE", "disable")

			if dbPassword == "" {
				return nil, fmt.Errorf("either DATABASE_URL or DB_PASSWORD must be set when USE_MONGO=false")
			}

			databaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
				dbUser, dbPassword, dbHost, dbPort, dbName, sslMode)
		}
	}

	cfg := &Config{
		UseMongo:       useMongo,
		DatabaseURL:    databaseURL,
		MongoURI:       mongoURI,
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
	if c.JWTSecret == "" {
		return fmt.Errorf("required environment variable JWT_SECRET is not set")
	}
	if c.GeminiAPIKey == "" {
		return fmt.Errorf("required environment variable GEMINI_API_KEY is not set")
	}
	if c.GoogleClientID == "" {
		return fmt.Errorf("required environment variable GOOGLE_CLIENT_ID is not set")
	}

	if c.UseMongo {
		if c.MongoURI == "" {
			return fmt.Errorf("required environment variable MONGODB_URI is not set when USE_MONGO=true")
		}
	} else {
		if c.DatabaseURL == "" {
			return fmt.Errorf("required environment variable DATABASE_URL is not set when USE_MONGO=false")
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
