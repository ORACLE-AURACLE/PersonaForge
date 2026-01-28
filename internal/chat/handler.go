package chat

import (
	"fmt"

	"github.com/PersonaForge/backend/internal/response"
	"github.com/gin-gonic/gin"
)

// Handler handles chat HTTP requests
type Handler struct {
	service *Service
}

// NewHandler creates a new chat handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// SendMessage godoc
// @Summary Send a chat message
// @Description Send a message to a persona and get a response
// @Tags chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SendMessageRequest true "Chat message"
// @Success 200 {object} response.APIResponse{data=SendMessageResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/chat [post]
func (h *Handler) SendMessage(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	// Get user info
	var userID *int
	isAuthenticated := false

	if id, exists := c.Get("user_id"); exists {
		uid := id.(int)
		userID = &uid
		isAuthenticated = true
	}

	if auth, exists := c.Get("is_authenticated"); exists {
		isAuthenticated = auth.(bool)
	}

	resp, err := h.service.SendMessage(c.Request.Context(), userID, isAuthenticated, req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// GetConversationHistory godoc
// @Summary Get conversation history
// @Description Retrieve the conversation history for the authenticated user's current session and a persona
// @Tags chat
// @Accept json
// @Produce json
// @Param persona_id query int true "Persona ID"
// @Success 200 {object} response.APIResponse{data=ConversationHistory}
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/chat/history [get]
func (h *Handler) GetConversationHistory(c *gin.Context) {
	personaIDStr := c.Query("persona_id")

	if personaIDStr == "" {
		response.BadRequest(c, "persona_id is required")
		return
	}

	// Auth middleware guarantees these
	userIDVal, ok := c.Get("user_id")
	if !ok {
		response.Unauthorized(c, "Authentication required")
		return
	}
	userID := userIDVal.(int)

	sessionIDVal, ok := c.Get("session_id")
	if !ok {
		response.Unauthorized(c, "Session not found")
		return
	}
	sessionID := sessionIDVal.(string)

	var personaID int
	if _, err := fmt.Sscanf(personaIDStr, "%d", &personaID); err != nil {
		response.BadRequest(c, "Invalid persona_id")
		return
	}

	history, err := h.service.GetConversationHistory(c.Request.Context(), userID, sessionID, personaID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, history)
}

// GenerateInsight godoc
// @Summary Generate conversation insights
// @Description Generate structured insights about the authenticated user's current session
// @Tags chat
// @Accept json
// @Produce json
// @Success 200 {object} response.APIResponse{data=InsightResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/insight [post]
func (h *Handler) GenerateInsight(c *gin.Context) {
	userIDVal, ok := c.Get("user_id")
	if !ok {
		response.Unauthorized(c, "Authentication required")
		return
	}
	userID := userIDVal.(int)

	sessionIDVal, ok := c.Get("session_id")
	if !ok {
		response.Unauthorized(c, "Session not found")
		return
	}
	sessionID := sessionIDVal.(string)

	insight, err := h.service.GenerateInsight(c.Request.Context(), userID, sessionID)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, insight)
}
