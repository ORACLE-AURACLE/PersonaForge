package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/PersonaForge/backend/internal/auth"
	"github.com/PersonaForge/backend/internal/gemini"
	"github.com/PersonaForge/backend/internal/persona"
	"github.com/PersonaForge/backend/internal/storage"
	"github.com/gin-gonic/gin"
)

type fakeChatRepo struct {
	nextSessionID int
	nextMsgID     int
	sessions      map[string]*storage.Session
	messages      map[int][]MessageDTO // key: session db id
}

func newFakeChatRepo() *fakeChatRepo {
	return &fakeChatRepo{
		nextSessionID: 1,
		nextMsgID:     1,
		sessions:      map[string]*storage.Session{},
		messages:      map[int][]MessageDTO{},
	}
}

func (r *fakeChatRepo) CreateSession(userID *int, sessionID string, isAnonymous bool, expiresAt time.Time) (int, error) {
	id := r.nextSessionID
	r.nextSessionID++
	r.sessions[sessionID] = &storage.Session{
		ID:          id,
		UserID:      userID,
		SessionID:   sessionID,
		IsAnonymous: isAnonymous,
		CreatedAt:   time.Now(),
		ExpiresAt:   expiresAt,
	}
	return id, nil
}

func (r *fakeChatRepo) GetSessionByID(sessionID string) (*storage.Session, error) {
	return r.sessions[sessionID], nil
}

func (r *fakeChatRepo) GetConversationHistory(sessionDBID int, personaID int) ([]MessageDTO, error) {
	var out []MessageDTO
	for _, m := range r.messages[sessionDBID] {
		if m.PersonaID == personaID {
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *fakeChatRepo) GetAllMessagesForSession(sessionDBID int) ([]MessageDTO, error) {
	return append([]MessageDTO(nil), r.messages[sessionDBID]...), nil
}

func (r *fakeChatRepo) SaveMessage(sessionDBID int, personaID int, role string, content string) (*MessageDTO, error) {
	m := MessageDTO{
		ID:        r.nextMsgID,
		SessionID: sessionDBID,
		PersonaID: personaID,
		Role:      role,
		Content:   content,
		CreatedAt: time.Now(),
	}
	r.nextMsgID++
	r.messages[sessionDBID] = append(r.messages[sessionDBID], m)
	return &m, nil
}

func (r *fakeChatRepo) SaveTokenUsage(sessionDBID int, promptTokens int, completionTokens int, totalTokens int) error {
	return nil
}

func (r *fakeChatRepo) MigrateSession(sessionID string, userID int) error { return nil }

type fakePersonaSvc struct{}

func (p *fakePersonaSvc) GetPersonaBlueprint(id int) (*persona.PersonaBlueprint, error) {
	return &persona.PersonaBlueprint{
		Name:        "Test Persona",
		Description: "desc",
		Personality: []string{"a"},
		Expertise:   []string{"e"},
		Tone:        "t",
		Guidelines:  []string{"g"},
	}, nil
}

type fakeGemini struct{}

func (g *fakeGemini) GenerateContent(ctx context.Context, req gemini.GenerateContentRequest) (*gemini.GenerateContentResponse, error) {
	return &gemini.GenerateContentResponse{
		Text:         "hello",
		PromptTokens: 1,
		OutputTokens: 1,
		TotalTokens:  2,
	}, nil
}

func (g *fakeGemini) GenerateInsight(ctx context.Context, messages []gemini.Message) (*gemini.InsightResponse, error) {
	return &gemini.InsightResponse{Analysis: "insight"}, nil
}

func TestChatEndpoints_GuestNoHistory_AuthHasHistoryAndInsight(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := newFakeChatRepo()
	svc := &Service{
		repo:           repo,
		geminiClient:   &fakeGemini{},
		personaService: &fakePersonaSvc{},
	}
	h := NewHandler(svc)
	jwtSvc := auth.NewJWTService("secret", 30)

	r := gin.New()
	api := r.Group("/api")
	RegisterRoutes(api, h, jwtSvc)

	// Guest can send message (creates session)
	var guestSession string
	{
		body := map[string]any{"persona_id": 1, "message": "hi"}
		b, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 got %d body=%s", w.Code, w.Body.String())
		}
		var parsed struct {
			Status string `json:"status"`
			Data   struct {
				SessionID string `json:"session_id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &parsed); err != nil {
			t.Fatalf("failed to unmarshal response: %v body=%s", err, w.Body.String())
		}
		guestSession = parsed.Data.SessionID
		if guestSession == "" {
			t.Fatalf("expected session_id in response body=%s", w.Body.String())
		}
	}

	// Guest cannot access history (auth required)
	{
		req := httptest.NewRequest(http.MethodGet, "/api/chat/history?persona_id=1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 got %d body=%s", w.Code, w.Body.String())
		}
	}

	// Seed an authenticated session + messages
	{
		uid := 1
		_, _ = repo.CreateSession(&uid, "authsess", false, time.Now().Add(1*time.Hour))
		sess := repo.sessions["authsess"]
		_, _ = repo.SaveMessage(sess.ID, 1, "user", "hi")
		_, _ = repo.SaveMessage(sess.ID, 1, "assistant", "yo")
	}

	// Auth history succeeds for own session
	{
		token, _ := jwtSvc.GenerateToken(1, "u@example.com", "authsess")
		req := httptest.NewRequest(http.MethodGet, "/api/chat/history?persona_id=1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 got %d body=%s", w.Code, w.Body.String())
		}
	}

	// Auth history fails if session ownership mismatches (session not found)
	{
		token, _ := jwtSvc.GenerateToken(2, "u2@example.com", "authsess")
		req := httptest.NewRequest(http.MethodGet, "/api/chat/history?persona_id=1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d body=%s", w.Code, w.Body.String())
		}
	}

	// Insight requires auth
	{
		req := httptest.NewRequest(http.MethodPost, "/api/insight", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 got %d body=%s", w.Code, w.Body.String())
		}
	}

	// Insight works for authenticated session
	{
		token, _ := jwtSvc.GenerateToken(1, "u@example.com", "authsess")
		req := httptest.NewRequest(http.MethodPost, "/api/insight", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 got %d body=%s", w.Code, w.Body.String())
		}
	}

	_ = guestSession
}
