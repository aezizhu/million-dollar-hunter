package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"time"
)

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func (s *Server) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if req.RefreshToken == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if s.RefreshTokens == nil {
		http.Error(w, "not implemented", http.StatusNotImplemented)
		return
	}
	expectedAud := os.Getenv("JWT_AUDIENCE")
	claims, err := s.JWT.ValidateToken(req.RefreshToken, expectedAud)
	if err != nil || claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if s.RefreshTokens == nil {
		http.Error(w, "not implemented", http.StatusNotImplemented)
		return
	}
	now := time.Now()
	if _, err := s.RefreshTokens.GetValidRefreshToken(r.Context(), req.RefreshToken, now); err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if err := s.RefreshTokens.RevokeRefreshToken(r.Context(), req.RefreshToken); err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	access, newRefresh, exp, err := s.JWT.GeneratePair(claims.Subject, claims.Email)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if _, err := s.RefreshTokens.CreateRefreshToken(r.Context(), claims.Subject, newRefresh, exp); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if s.Audit != nil {
		_ = s.Audit.Log(r.Context(), &claims.Subject, "refresh_success", nil, nil)
	}
	_ = json.NewEncoder(w).Encode(RefreshResponse{
		AccessToken:  access,
		RefreshToken: newRefresh,
		ExpiresIn:    exp.Unix() - time.Now().Unix(),
	})
}

func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {
	if s.Audit != nil {
		if claims := ClaimsFromContext(r); claims != nil {
			_ = s.Audit.Log(r.Context(), &claims.Subject, "logout", nil, nil)
			if s.RefreshTokens != nil {
				_ = s.RefreshTokens.RevokeAllForUser(r.Context(), claims.Subject)
			}
		}
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ok":true}`))
}
