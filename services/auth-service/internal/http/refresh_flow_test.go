package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	jwtmgr "github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/jwt"
	"github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/store"
)

type memRefresh struct {
	valid map[string]time.Time
	user  map[string]string
}

func (m *memRefresh) CreateRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) (store.RefreshToken, error) {
	if m.valid == nil {
		m.valid = map[string]time.Time{}
	}
	if m.user == nil {
		m.user = map[string]string{}
	}
	m.valid[token] = expiresAt
	m.user[token] = userID
	return store.RefreshToken{Token: token, UserID: userID, ExpiresAt: expiresAt}, nil
}
func (m *memRefresh) GetValidRefreshToken(ctx context.Context, token string, now time.Time) (store.RefreshToken, error) {
	if m.valid == nil {
		return store.RefreshToken{}, assertErr{}
	}
	exp, ok := m.valid[token]
	if !ok || now.After(exp) {
		return store.RefreshToken{}, assertErr{}
	}
	return store.RefreshToken{Token: token, UserID: m.user[token], ExpiresAt: exp}, nil
}
func (m *memRefresh) RevokeRefreshToken(ctx context.Context, token string) error {
	if m.valid == nil {
		m.valid = map[string]time.Time{}
	}
	m.valid[token] = time.Now().Add(-1 * time.Minute)
	delete(m.valid, token)
	delete(m.user, token)
	return nil
}
func (m *memRefresh) RevokeAllForUser(ctx context.Context, userID string) error {
	for t, uid := range m.user {
		if uid == userID {
			delete(m.valid, t)
			delete(m.user, t)
		}
	}
	return nil
}

type memAudit struct {
	events []string
}

func (m *memAudit) Log(ctx context.Context, userID *string, event string, ip *string, meta *string) error {
	m.events = append(m.events, event)
	return nil
}
func (m *memAudit) CountRecentLoginFailures(ctx context.Context, userID *string, window time.Duration) (int, error) {
	return 0, nil
}

type jwtFixed struct {
	j *jwtmgr.Manager
}

func (f *jwtFixed) GeneratePair(userID, email string) (string, string, time.Time, error) {
	return f.j.GeneratePair(userID, email)
}
func (f *jwtFixed) ValidateToken(tokenStr string, expectedAud string) (*jwtmgr.Claims, error) {
	return f.j.ValidateToken(tokenStr, expectedAud)
}

func TestRefreshSuccess_RotatesAndPersists(t *testing.T) {
	_ = os.Setenv("JWT_SIGNING_KEY", "test-secret")
	_ = os.Setenv("JWT_ISSUER", "mdh-auth")
	_ = os.Setenv("JWT_AUDIENCE", "mdh-api")
	m := jwtmgr.New("mdh-auth", "mdh-api", 1*time.Minute, 5*time.Minute, []byte("test-secret"))

	access, refresh, _, err := m.GeneratePair("u-1", "u@example.com")
	if err != nil || access == "" || refresh == "" {
		t.Fatalf("failed to generate tokens")
	}

	mem := &memRefresh{}
	_, _ = mem.CreateRefreshToken(context.Background(), "u-1", refresh, time.Now().Add(5*time.Minute))

	s := &Server{
		JWT:           &jwtFixed{j: m},
		RefreshTokens: mem,
		Audit:         &memAudit{},
	}
	body, _ := json.Marshal(RefreshRequest{RefreshToken: refresh})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.Refresh(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp RefreshResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp.AccessToken == "" || resp.RefreshToken == "" || resp.ExpiresIn <= 0 {
		t.Fatalf("bad response")
	}
	if _, err := mem.GetValidRefreshToken(context.Background(), refresh, time.Now()); err == nil {
		t.Fatalf("old token not revoked")
	}
	if _, err := mem.GetValidRefreshToken(context.Background(), resp.RefreshToken, time.Now()); err != nil {
		t.Fatalf("new refresh not persisted")
	}
}

func TestRefreshUnauthorized_InvalidOrMissing(t *testing.T) {
	_ = os.Setenv("JWT_SIGNING_KEY", "test-secret")
	_ = os.Setenv("JWT_ISSUER", "mdh-auth")
	_ = os.Setenv("JWT_AUDIENCE", "mdh-api")
	m := jwtmgr.New("mdh-auth", "mdh-api", 1*time.Minute, 5*time.Minute, []byte("test-secret"))

	s := &Server{
		JWT:           &jwtFixed{j: m},
		RefreshTokens: &memRefresh{},
		Audit:         &memAudit{},
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader([]byte(`{"refresh_token":""}`)))
	w := httptest.NewRecorder()
	s.Refresh(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token")
	}

	_, bogus, _, _ := m.GeneratePair("u-1", "u@example.com")
	body, _ := json.Marshal(RefreshRequest{RefreshToken: bogus})
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(body))
	w2 := httptest.NewRecorder()
	s.Refresh(w2, req2)
	if w2.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for nonexistent token, got %d", w2.Code)
	}
}
