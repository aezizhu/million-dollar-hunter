package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
)

type RotateRequest struct {
	Bits int `json:"bits"`
}

type RotateResponse struct {
	KID string `json:"kid"`
}

type KeyRotator interface {
	Rotate(bits int) (string, error)
}

func (s *Server) RotateKeys(w http.ResponseWriter, r *http.Request) {
	admin := os.Getenv("JWT_ROTATE_ADMIN_TOKEN")
	if admin == "" || r.Header.Get("X-Admin-Token") != admin {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req RotateRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Bits <= 0 {
		req.Bits = 2048
	}
	if rot, ok := any(s.JWT).(KeyRotator); ok {
		kid, err := rot.Rotate(req.Bits)
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(RotateResponse{KID: kid})
		return
	}
	http.Error(w, "server error", http.StatusInternalServerError)
}

func JWKSCacheTTL() int {
	if v := os.Getenv("JWKS_CACHE_TTL_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 300
}
