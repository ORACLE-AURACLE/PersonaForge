package chat

import (
	"time"

	"github.com/PersonaForge/backend/internal/auth"
	"github.com/PersonaForge/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers chat routes
func RegisterRoutes(router *gin.RouterGroup, handler *Handler, jwtService *auth.JWTService) {
	// Rate limiting for chat endpoints
	chatRateLimit := middleware.AdaptiveRateLimit(middleware.RateLimitConfig{
		FreeLimit: 10, // 10 requests per minute for free users
		AuthLimit: 30, // 30 requests per minute for authenticated users
		Window:    1 * time.Minute,
	})

	chat := router.Group("/chat")
	{
		chat.POST("", middleware.OptionalAuthMiddleware(jwtService), chatRateLimit, handler.SendMessage)
		// History for the authenticated user's current session.
		chat.GET("/history", middleware.AuthMiddleware(jwtService), handler.GetConversationHistory)
		// History by session id for both guests (anonymous sessions) and authenticated users.
		chat.GET("/:session_id/history", middleware.OptionalAuthMiddleware(jwtService), chatRateLimit, handler.GetConversationHistoryBySession)
	}

	// Insight: open to all (guests and authenticated). GET, query session_id and optional persona_id.
	router.GET("/insight", middleware.OptionalAuthMiddleware(jwtService), chatRateLimit, handler.GenerateInsight)
}
