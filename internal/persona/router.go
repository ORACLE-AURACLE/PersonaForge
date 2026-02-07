package persona

import (
	"github.com/PersonaForge/backend/internal/auth"
	"github.com/PersonaForge/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers persona routes
func RegisterRoutes(router *gin.RouterGroup, handler *Handler, jwtService *auth.JWTService) {
	personas := router.Group("/personas")
	{
		// Public routes (with optional auth)
		personas.GET("", middleware.OptionalAuthMiddleware(jwtService), handler.ListPersonas)
		personas.GET("/by-session", handler.ListPersonasBySession) // no auth; session_id query or X-Session-ID header
		personas.GET("/:id", handler.GetPersona)

		// Protected routes (require auth)
		personas.POST("", middleware.OptionalAuthMiddleware(jwtService), handler.CreatePersona)
		personas.DELETE("/:id", middleware.AuthMiddleware(jwtService), handler.DeletePersona)
	}
}
