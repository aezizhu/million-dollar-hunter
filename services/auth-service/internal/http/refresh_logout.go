package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/hlog"
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
	log := hlog.FromRequest(r)
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	http.Error(w, "not implemented", http.StatusNotImplemented)
	log.Info().Msg("refresh not implemented in MVP")
}

func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ok":true}`))
}
