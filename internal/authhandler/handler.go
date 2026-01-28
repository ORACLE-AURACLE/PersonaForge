package authhandler

import (
	"github.com/PersonaForge/backend/internal/response"
	"github.com/gin-gonic/gin"
)

// Handler handles authentication HTTP requests
type Handler struct {
	service *Service
}

// NewHandler creates a new auth handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GoogleLogin godoc
// @Summary Login with Google
// @Description Authenticate using Google ID token and optionally migrate an anonymous session
// @Tags auth
// @Accept json
// @Produce json
// @Param request body GoogleLoginRequest true "Google login request"
// @Success 200 {object} response.APIResponse{data=AuthResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/auth/google [post]
func (h *Handler) GoogleLogin(c *gin.Context) {
	var req GoogleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	resp, err := h.service.GoogleLogin(c.Request.Context(), req.IDToken, req.SessionID)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// CreateAnonymousSession godoc
// @Summary Create anonymous session
// @Description Create a new anonymous session for free mode usage
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} response.APIResponse{data=AnonymousSessionResponse}
// @Failure 500 {object} response.APIResponse
// @Router /api/auth/anonymous [post]
func (h *Handler) CreateAnonymousSession(c *gin.Context) {
	resp, err := h.service.CreateAnonymousSession()
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, resp)
}
