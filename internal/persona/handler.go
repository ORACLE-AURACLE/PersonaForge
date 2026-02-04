package persona

import (
	"strconv"

	"github.com/PersonaForge/backend/internal/response"
	"github.com/gin-gonic/gin"
)

// Handler handles persona HTTP requests
type Handler struct {
	service *Service
}

// NewHandler creates a new persona handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ListPersonas godoc
// @Summary List all personas
// @Description Get all personas available to the user (4 defaults + custom personas). Optionally provide X-Session-ID header to retrieve guest personas.
// @Tags personas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Session-ID header string false "Session ID for guest users (optional)"
// @Success 200 {object} response.APIResponse{data=[]PersonaResponse}
// @Failure 500 {object} response.APIResponse
// @Router /api/personas [get]
func (h *Handler) ListPersonas(c *gin.Context) {
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

	// Optional for guests: allow retrieving their one custom persona if they provide a session id
	sessionIDHeader := c.GetHeader("X-Session-ID")
	var sessionID *string
	if sessionIDHeader != "" {
		sessionID = &sessionIDHeader
	}

	personas, err := h.service.ListPersonas(userID, sessionID, isAuthenticated)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve personas")
		return
	}

	response.Success(c, personas)
}

// GetPersona godoc
// @Summary Get a persona by ID
// @Description Retrieve details of a specific persona
// @Tags personas
// @Accept json
// @Produce json
// @Param id path int true "Persona ID"
// @Success 200 {object} response.APIResponse{data=PersonaResponse}
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/personas/{id} [get]
func (h *Handler) GetPersona(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid persona ID")
		return
	}

	persona, err := h.service.GetPersona(id)
	if err != nil {
		response.NotFound(c, "Persona not found")
		return
	}

	response.Success(c, persona)
}

// CreatePersona godoc
// @Summary Create a custom persona
// @Description Create a new custom persona (limited to 2 for free users). Authenticated users use Bearer token. Guests must provide X-Session-ID header from /api/auth/anonymous.
// @Tags personas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param X-Session-ID header string false "Session ID for guest users (required if not authenticated)"
// @Param request body CreatePersonaRequest true "Persona details"
// @Success 201 {object} response.APIResponse{data=PersonaResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 403 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /api/personas [post]
func (h *Handler) CreatePersona(c *gin.Context) {
	var req CreatePersonaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	// Get user info
	var userID *int
	var sessionID *string
	isAuthenticated := false

	if id, exists := c.Get("user_id"); exists {
		uid := id.(int)
		userID = &uid
		isAuthenticated = true
	}

	if auth, exists := c.Get("is_authenticated"); exists {
		isAuthenticated = auth.(bool)
	}

	// Guests must provide a session id in header (issued by /api/auth/anonymous)
	if !isAuthenticated {
		sid := c.GetHeader("X-Session-ID")
		if sid == "" {
			response.BadRequest(c, "X-Session-ID header required for guest persona creation")
			return
		}
		sessionID = &sid
	}

	persona, err := h.service.CreateCustomPersona(userID, sessionID, isAuthenticated, req)
	if err != nil {
		if err.Error() == "free users can only create 2 custom personas" {
			response.Forbidden(c, err.Error())
			return
		}
		if err.Error() == "invalid or expired session" || err.Error() == "session_id is required for guest persona creation" {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to create persona")
		return
	}

	response.Created(c, persona)
}

// DeletePersona godoc
// @Summary Delete a custom persona
// @Description Delete a user's custom persona (cannot delete default personas)
// @Tags personas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Persona ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/personas/{id} [delete]
func (h *Handler) DeletePersona(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid persona ID")
		return
	}

	// Get user ID
	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}
	userID := userIDVal.(int)

	if err := h.service.DeletePersona(id, userID); err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "Persona deleted successfully"})
}
