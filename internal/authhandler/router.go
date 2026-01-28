package authhandler

import "github.com/gin-gonic/gin"

// RegisterRoutes registers authentication routes
func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	auth := router.Group("/auth")
	{
		auth.POST("/google", handler.GoogleLogin)
		auth.POST("/anonymous", handler.CreateAnonymousSession)
	}
}
