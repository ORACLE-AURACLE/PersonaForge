package persona

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/PersonaForge/backend/internal/auth"
	"github.com/PersonaForge/backend/internal/storage"
	"github.com/gin-gonic/gin"
)

type fakePersonaRepo struct {
	nextID       int
	personas     []storage.Persona
	activeSesion map[string]bool
}

func newFakePersonaRepo() *fakePersonaRepo {
	r := &fakePersonaRepo{
		nextID:       1,
		activeSesion: map[string]bool{},
	}
	// Seed defaults
	for _, bp := range DefaultPersonas() {
		bpJSON, _ := MarshalBlueprint(bp)
		r.personas = append(r.personas, storage.Persona{
			ID:        r.nextID,
			UserID:    nil,
			SessionID: nil,
			Name:      bp.Name,
			Blueprint: bpJSON,
			IsDefault: true,
			CreatedAt: time.Now(),
		})
		r.nextID++
	}
	return r
}

func (r *fakePersonaRepo) SessionActive(sessionID string) (bool, error) {
	return r.activeSesion[sessionID], nil
}

func (r *fakePersonaRepo) CountCustomPersonasForSession(sessionID string) (int, error) {
	n := 0
	for _, p := range r.personas {
		if !p.IsDefault && p.SessionID != nil && *p.SessionID == sessionID {
			n++
		}
	}
	return n, nil
}

func (r *fakePersonaRepo) CreatePersona(userID *int, sessionID *string, name string, blueprint string) (*storage.Persona, error) {
	p := storage.Persona{
		ID:        r.nextID,
		UserID:    userID,
		SessionID: sessionID,
		Name:      name,
		Blueprint: blueprint,
		IsDefault: false,
		CreatedAt: time.Now(),
	}
	r.nextID++
	r.personas = append(r.personas, p)
	return &p, nil
}

func (r *fakePersonaRepo) GetPersonaByID(id int) (*storage.Persona, error) {
	for _, p := range r.personas {
		if p.ID == id {
			cp := p
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *fakePersonaRepo) ListPersonasForUser(userID int) ([]storage.Persona, error) {
	var out []storage.Persona
	for _, p := range r.personas {
		if p.IsDefault || (p.UserID != nil && *p.UserID == userID) {
			out = append(out, p)
		}
	}
	return out, nil
}

func (r *fakePersonaRepo) ListPersonasForSession(sessionID string) ([]storage.Persona, error) {
	var out []storage.Persona
	for _, p := range r.personas {
		if p.IsDefault || (p.SessionID != nil && *p.SessionID == sessionID) {
			out = append(out, p)
		}
	}
	return out, nil
}

func (r *fakePersonaRepo) ListDefaultPersonas() ([]storage.Persona, error) {
	var out []storage.Persona
	for _, p := range r.personas {
		if p.IsDefault {
			out = append(out, p)
		}
	}
	return out, nil
}

func (r *fakePersonaRepo) DeletePersona(id int, userID int) error { return nil }
func (r *fakePersonaRepo) InitializeDefaultPersonas() error       { return nil }

func TestPersonaEndpoints_GuestLimitAndAuthUnlimited(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := newFakePersonaRepo()
	repo.activeSesion["guest-session"] = true

	svc := NewService(repo)
	h := NewHandler(svc)
	jwtSvc := auth.NewJWTService("secret", 30)

	r := gin.New()
	api := r.Group("/api")
	RegisterRoutes(api, h, jwtSvc)

	// Guest list defaults only
	{
		req := httptest.NewRequest(http.MethodGet, "/api/personas", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 got %d body=%s", w.Code, w.Body.String())
		}
	}

	// Guest create without X-Session-ID should fail
	{
		body := `{"name":"G1","blueprint":{"name":"G1","description":"d","personality":["a"],"expertise":["e"],"tone":"t","guidelines":["g"]}}`
		req := httptest.NewRequest(http.MethodPost, "/api/personas", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d body=%s", w.Code, w.Body.String())
		}
	}

	// Guest create with X-Session-ID succeeds once
	{
		body := `{"name":"G1","blueprint":{"name":"G1","description":"d","personality":["a"],"expertise":["e"],"tone":"t","guidelines":["g"]}}`
		req := httptest.NewRequest(http.MethodPost, "/api/personas", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Session-ID", "guest-session")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201 got %d body=%s", w.Code, w.Body.String())
		}
	}
	// Second guest persona should be forbidden
	{
		body := `{"name":"G2","blueprint":{"name":"G2","description":"d","personality":["a"],"expertise":["e"],"tone":"t","guidelines":["g"]}}`
		req := httptest.NewRequest(http.MethodPost, "/api/personas", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Session-ID", "guest-session")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403 got %d body=%s", w.Code, w.Body.String())
		}
	}

	// Auth user can create multiple (no session header required)
	{
		token, _ := jwtSvc.GenerateToken(123, "u@example.com", "sess")
		for i := 0; i < 2; i++ {
			reqBody := map[string]any{
				"name": "U",
				"blueprint": map[string]any{
					"name":        "U",
					"description": "d",
					"personality": []string{"a"},
					"expertise":   []string{"e"},
					"tone":        "t",
					"guidelines":  []string{"g"},
				},
			}
			b, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/api/personas", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusCreated {
				t.Fatalf("expected 201 got %d body=%s", w.Code, w.Body.String())
			}
		}
	}
}


