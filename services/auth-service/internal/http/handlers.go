package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"

	"github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/auth"
	jwtmgr "github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/jwt"
	"github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/store"
)

type JWTManager interface {
	GeneratePair(userID, email string) (string, string, time.Time, error)
}

type Server struct {
	Logger *zerolog.Logger
	JWT    JWTManager
	Store  store.UserStore
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	UserID string `json:"user_id"`
}

func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	log := hlog.FromRequest(r)
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if os.Getenv("ENABLE_MULTI_USER") != "true" {
		if req.Username != "aezi" || req.Password != "Aa@123456789" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		userID, email := "00000000-0000-0000-0000-000000000001", "owner@example.com"
		access, refresh, exp, err := s.JWT.GeneratePair(userID, email)
		if err != nil {
			log.Error().Err(err).Msg("generate tokens failed")
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(LoginResponse{
			AccessToken:  access,
			RefreshToken: refresh,
			ExpiresIn:    exp.Unix() - time.Now().Unix(),
		})
		return
	}
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (s *Server) Register(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("ENABLE_MULTI_USER") != "true" {
		http.Error(w, "not implemented", http.StatusNotImplemented)
		return
	}
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if !strings.Contains(req.Email, "@") {
		http.Error(w, "invalid email", http.StatusBadRequest)
		return
	}
	if err := auth.ValidatePasswordPolicy(req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	u, err := s.Store.Create(r.Context(), req.Email, hash)
	if err != nil {
		http.Error(w, "conflict", http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(RegisterResponse{UserID: u.ID})
}

func (s *Server) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ok":true}`))
}

var _ = jwtmgr.New
