package authhandler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/PersonaForge/backend/internal/auth"
	"github.com/PersonaForge/backend/internal/storage"
	"github.com/gin-gonic/gin"
)

type fakeAuthRepo struct {
	user *storage.User
}

func (r *fakeAuthRepo) GetUserByGoogleID(googleID string) (*storage.User, error) { return r.user, nil }
func (r *fakeAuthRepo) CreateUser(googleID string, email string) (*storage.User, error) {
	r.user = &storage.User{ID: 1, GoogleID: googleID, Email: email, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	return r.user, nil
}

type fakeGoogle struct {
	info *auth.GoogleUserInfo
	err  error
}

func (g *fakeGoogle) VerifyIDToken(ctx context.Context, idToken string) (*auth.GoogleUserInfo, error) {
	return g.info, g.err
}

type fakeJWT struct{}

func (j *fakeJWT) GenerateToken(userID int, email, sessionID string) (string, error) { return "jwt-token", nil }
func (j *fakeJWT) ExpiryMinutes() int                                              { return 30 }

type fakeChatRepo struct {
	created bool
}

func (c *fakeChatRepo) CreateSession(userID *int, sessionID string, isAnonymous bool, expiresAt time.Time) (int, error) {
	c.created = true
	return 1, nil
}
func (c *fakeChatRepo) MigrateSession(sessionID string, userID int) error { return nil }

func TestAuthEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &Service{
		repo:       &fakeAuthRepo{},
		googleAuth: &fakeGoogle{info: &auth.GoogleUserInfo{GoogleID: "gid", Email: "u@example.com", Name: "U"}},
		jwtService: &fakeJWT{},
		chatRepo:   &fakeChatRepo{},
	}
	h := NewHandler(svc)

	r := gin.New()
	api := r.Group("/api")
	RegisterRoutes(api, h)

	// anonymous session
	{
		req := httptest.NewRequest(http.MethodPost, "/api/auth/anonymous", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 got %d body=%s", w.Code, w.Body.String())
		}
	}

	// google login invalid body
	{
		req := httptest.NewRequest(http.MethodPost, "/api/auth/google", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d body=%s", w.Code, w.Body.String())
		}
	}

	// google login ok
	{
		b, _ := json.Marshal(map[string]string{"id_token": "token"})
		req := httptest.NewRequest(http.MethodPost, "/api/auth/google", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 got %d body=%s", w.Code, w.Body.String())
		}
	}
}


