package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/keystore"
)

type JWKSResponse struct {
	Keys []keystore.PublicKey `json:"keys"`
}

func (s *Server) JWKS(w http.ResponseWriter, r *http.Request) {
	type keyLister interface{ ListPublic() []keystore.PublicKey }
	if jl, ok := any(s.JWT).(keyLister); ok {
		keys := jl.ListPublic()
		resp := JWKSResponse{Keys: keys}
		data, _ := json.Marshal(resp)
		sum := sha256.Sum256(data)
		etag := `W/"` + hex.EncodeToString(sum[:]) + `"`
		if match := r.Header.Get("If-None-Match"); match != "" && match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		cacheSec := 300
		shortTTL := 60
		if v := os.Getenv("JWKS_CACHE_TTL_SECONDS"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				cacheSec = n
			}
		}
		if v := os.Getenv("JWKS_ROTATION_TTL_SECONDS"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				shortTTL = n
			}
		}
		short := false
		for _, k := range keys {
			_ = k
			if len(keys) > 1 {
				short = true
				break
			}
		}
		ttl := cacheSec
		if short {
			ttl = shortTTL
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(ttl))
		w.Header().Set("ETag", etag)
		_, _ = w.Write(data)
		return
	}
	http.Error(w, "server error", http.StatusInternalServerError)
}
