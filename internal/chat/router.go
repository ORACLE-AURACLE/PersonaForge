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
		// History is restricted to authenticated users only.
		chat.GET("/history", middleware.AuthMiddleware(jwtService), handler.GetConversationHistory)
	}

	// Insight endpoint
	// Insights require authentication (guests should not be able to retrieve post-hoc summaries).
	router.POST("/insight", middleware.AuthMiddleware(jwtService), handler.GenerateInsight)
}
