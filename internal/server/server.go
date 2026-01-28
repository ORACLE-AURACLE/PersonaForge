package server

import (
	"context"
	"fmt"

	"github.com/PersonaForge/backend/internal/auth"
	"github.com/PersonaForge/backend/internal/authhandler"
	"github.com/PersonaForge/backend/internal/chat"
	"github.com/PersonaForge/backend/internal/config"
	"github.com/PersonaForge/backend/internal/gemini"
	"github.com/PersonaForge/backend/internal/middleware"
	"github.com/PersonaForge/backend/internal/persona"
	"github.com/PersonaForge/backend/internal/response"
	"github.com/PersonaForge/backend/internal/storage"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Server represents the HTTP server
type Server struct {
	router *gin.Engine
	config *config.Config
	db     *storage.Database
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config) (*Server, error) {
	// Initialize database
	db, err := storage.NewDatabase(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	return &Server{
		router: router,
		config: cfg,
		db:     db,
	}, nil
}

// Setup initializes all routes and middleware
func (s *Server) Setup(ctx context.Context) error {
	// Apply global middleware
	s.router.Use(middleware.SecurityHeaders())

	// Initialize services
	jwtService := auth.NewJWTService(s.config.JWTSecret, s.config.JWTExpiryMinutes)
	googleAuthService := auth.NewGoogleAuthService(s.config.GoogleClientID)

	// Initialize Gemini client
	geminiClient, err := gemini.NewClient(ctx, s.config.GeminiAPIKey, s.config.GeminiModel)
	if err != nil {
		return fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Initialize repositories
	personaRepo := persona.NewRepository(s.db.DB)
	chatRepo := chat.NewRepository(s.db.DB)
	authRepo := authhandler.NewRepository(s.db.DB)

	// Initialize default personas
	personaService := persona.NewService(personaRepo)
	if err := personaService.InitializeDefaults(); err != nil {
		return fmt.Errorf("failed to initialize default personas: %w", err)
	}

	// Initialize services
	chatService := chat.NewService(chatRepo, personaRepo, geminiClient, personaService)
	authService := authhandler.NewService(authRepo, googleAuthService, jwtService, chatRepo)

	// Initialize handlers
	personaHandler := persona.NewHandler(personaService)
	chatHandler := chat.NewHandler(chatService)
	authHandler := authhandler.NewHandler(authService)

	// API routes
	api := s.router.Group("/api")
	{
		// Register routes
		authhandler.RegisterRoutes(api, authHandler)
		persona.RegisterRoutes(api, personaHandler, jwtService)
		chat.RegisterRoutes(api, chatHandler, jwtService)
	}

	// Swagger documentation at /docs
	s.router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	s.router.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "healthy"})
	})

	return nil
}

// Run starts the HTTP server
func (s *Server) Run() error {
	addr := ":" + s.config.Port
	fmt.Printf("Server starting on %s\n", addr)
	fmt.Printf("API documentation available at http://localhost%s/docs/index.html\n", addr)
	return s.router.Run(addr)
}

// Close closes the server and its dependencies
func (s *Server) Close() error {
	return s.db.Close()
}
