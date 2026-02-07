package server

import (
	"context"
	"fmt"
	"time"

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
	db     storage.DatabaseInterface
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config) (*Server, error) {
	var db storage.DatabaseInterface

	// Initialize database based on USE_MONGO config
	if cfg.UseMongo {
		mongoDB, err := storage.NewMongoDatabase(cfg.MongoURI)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
		}
		db = mongoDB
		fmt.Println("Using MongoDB database")
	} else {
		postgresDB, err := storage.NewPostgresDatabase(cfg.DatabaseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
		}
		db = postgresDB
		fmt.Println("Using PostgreSQL database")
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
	s.router.Use(middleware.CORSFromOrigins(s.config.CORSOrigins))
	s.router.Use(middleware.SecurityHeaders())

	// Initialize services
	jwtService := auth.NewJWTService(s.config.JWTSecret, s.config.JWTExpiryMinutes)
	googleAuthService := auth.NewGoogleAuthService(s.config.GoogleClientID)

	// Initialize Gemini client
	geminiClient, err := gemini.NewClient(ctx, s.config.GeminiAPIKey, s.config.GeminiModel)
	if err != nil {
		return fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Initialize repositories based on database type
	var personaRepoInstance interface{}
	var chatRepoInstance interface{}
	var authRepoInstance interface{}

	if s.config.UseMongo {
		mongoDBInstance, ok := s.db.(*storage.MongoDatabase)
		if !ok {
			return fmt.Errorf("failed to get MongoDB instance")
		}
		personaRepoInstance = persona.NewMongoRepository(mongoDBInstance)
		chatRepoInstance = chat.NewMongoRepository(mongoDBInstance)
		authRepoInstance = authhandler.NewMongoRepository(mongoDBInstance)
	} else {
		sqlDB := s.db.GetSQLDB()
		if sqlDB == nil {
			return fmt.Errorf("SQL database not available")
		}
		personaRepoInstance = persona.NewRepository(sqlDB)
		chatRepoInstance = chat.NewRepository(sqlDB)
		authRepoInstance = authhandler.NewRepository(sqlDB)
	}

	// Services use interfaces, so we can pass the concrete types directly
	// They will satisfy the interface requirements
	personaRepo := personaRepoInstance.(interface {
		SessionActive(sessionID string) (bool, error)
		CountCustomPersonasForSession(sessionID string) (int, error)
		CreatePersona(userID *int, sessionID *string, name string, blueprint string) (*storage.Persona, error)
		GetPersonaByID(id int) (*storage.Persona, error)
		ListPersonasForUser(userID int) ([]storage.Persona, error)
		ListPersonasForSession(sessionID string) ([]storage.Persona, error)
		ListCustomPersonasForSession(sessionID string) ([]storage.Persona, error)
		ListDefaultPersonas() ([]storage.Persona, error)
		DeletePersona(id int, userID int) error
		InitializeDefaultPersonas() error
	})

	chatRepo := chatRepoInstance.(interface {
		CreateSession(userID *int, sessionID string, isAnonymous bool, expiresAt time.Time) (int, error)
		GetSessionByID(sessionID string) (*storage.Session, error)
		GetConversationHistory(sessionDBID int, personaID int) ([]chat.MessageDTO, error)
		GetAllMessagesForSession(sessionDBID int) ([]chat.MessageDTO, error)
		SaveMessage(sessionDBID int, personaID int, role string, content string) (*chat.MessageDTO, error)
		SaveTokenUsage(sessionDBID int, promptTokens int, completionTokens int, totalTokens int) error
		MigrateSession(sessionID string, userID int) error
	})

	authRepo := authRepoInstance.(interface {
		GetUserByGoogleID(googleID string) (*storage.User, error)
		CreateUser(googleID string, email string) (*storage.User, error)
	})

	chatSessionRepo := chatRepoInstance.(interface {
		CreateSession(userID *int, sessionID string, isAnonymous bool, expiresAt time.Time) (int, error)
		MigrateSession(sessionID string, userID int) error
	})

	// Initialize default personas
	personaService := persona.NewService(personaRepo)
	if err := personaService.InitializeDefaults(); err != nil {
		return fmt.Errorf("failed to initialize default personas: %w", err)
	}

	// Initialize services
	chatService := chat.NewService(chatRepo, nil, geminiClient, personaService)
	authService := authhandler.NewService(authRepo, googleAuthService, jwtService, chatSessionRepo)

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
