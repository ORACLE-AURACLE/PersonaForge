package persona

import (
	"fmt"

	"github.com/PersonaForge/backend/internal/storage"
)

type personaRepo interface {
	SessionActive(sessionID string) (bool, error)
	CountCustomPersonasForSession(sessionID string) (int, error)
	CreatePersona(userID *int, sessionID *string, name string, blueprint string) (*storage.Persona, error)

	GetPersonaByID(id int) (*storage.Persona, error)
	ListPersonasForUser(userID int) ([]storage.Persona, error)
	ListPersonasForSession(sessionID string) ([]storage.Persona, error)
	ListDefaultPersonas() ([]storage.Persona, error)
	DeletePersona(id int, userID int) error
	InitializeDefaultPersonas() error
}

// Service handles persona business logic
type Service struct {
	repo personaRepo
}

// NewService creates a new persona service
func NewService(repo personaRepo) *Service {
	return &Service{repo: repo}
}

// CreateCustomPersona creates a new custom persona with limit enforcement
func (s *Service) CreateCustomPersona(userID *int, sessionID *string, isAuthenticated bool, req CreatePersonaRequest) (*PersonaResponse, error) {
	// Enforce limits:
	// - Guests: 2 custom personas per session
	// - Authenticated users: unlimited (for now)
	if !isAuthenticated {
		if sessionID == nil || *sessionID == "" {
			return nil, fmt.Errorf("session_id is required for guest persona creation")
		}

		active, err := s.repo.SessionActive(*sessionID)
		if err != nil {
			return nil, err
		}
		if !active {
			return nil, fmt.Errorf("invalid or expired session")
		}

		count, err := s.repo.CountCustomPersonasForSession(*sessionID)
		if err != nil {
			return nil, err
		}
		if count >= 2 {
			return nil, fmt.Errorf("free users can only create 2 custom personas")
		}
	} else {
		if userID == nil {
			return nil, fmt.Errorf("user_id is required for authenticated persona creation")
		}
	}

	// Marshal blueprint to JSON
	blueprintJSON, err := MarshalBlueprint(req.Blueprint)
	if err != nil {
		return nil, fmt.Errorf("invalid persona blueprint: %w", err)
	}

	// Create persona
	persona, err := s.repo.CreatePersona(userID, sessionID, req.Name, blueprintJSON)
	if err != nil {
		return nil, err
	}

	return s.toResponse(persona)
}

// GetPersona retrieves a persona by ID
func (s *Service) GetPersona(id int) (*PersonaResponse, error) {
	persona, err := s.repo.GetPersonaByID(id)
	if err != nil {
		return nil, err
	}

	return s.toResponse(persona)
}

// ListPersonas retrieves all personas for a user
func (s *Service) ListPersonas(userID *int, sessionID *string, isAuthenticated bool) ([]PersonaResponse, error) {
	var personas []storage.Persona
	var err error

	if isAuthenticated {
		if userID == nil {
			return nil, fmt.Errorf("user_id is required")
		}
		personas, err = s.repo.ListPersonasForUser(*userID)
	} else if sessionID != nil && *sessionID != "" {
		personas, err = s.repo.ListPersonasForSession(*sessionID)
	} else {
		// Guest without session: defaults only
		personas, err = s.repo.ListDefaultPersonas()
	}
	if err != nil {
		return nil, err
	}

	var responses []PersonaResponse
	for _, persona := range personas {
		resp, err := s.toResponse(&persona)
		if err != nil {
			return nil, err
		}
		responses = append(responses, *resp)
	}

	return responses, nil
}

// DeletePersona deletes a custom persona
func (s *Service) DeletePersona(id int, userID int) error {
	return s.repo.DeletePersona(id, userID)
}

// GetPersonaBlueprint retrieves and unmarshals a persona blueprint
func (s *Service) GetPersonaBlueprint(id int) (*PersonaBlueprint, error) {
	persona, err := s.repo.GetPersonaByID(id)
	if err != nil {
		return nil, err
	}

	return UnmarshalBlueprint(persona.Blueprint)
}

// InitializeDefaults creates default personas if they don't exist
func (s *Service) InitializeDefaults() error {
	return s.repo.InitializeDefaultPersonas()
}

func (s *Service) toResponse(persona *storage.Persona) (*PersonaResponse, error) {
	blueprint, err := UnmarshalBlueprint(persona.Blueprint)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal blueprint: %w", err)
	}

	return &PersonaResponse{
		ID:        persona.ID,
		Name:      persona.Name,
		Blueprint: *blueprint,
		IsDefault: persona.IsDefault,
		CreatedAt: persona.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}
