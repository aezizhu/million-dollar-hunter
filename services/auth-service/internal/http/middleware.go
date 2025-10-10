package httpapi

import (
	"context"
	"net/http"
	"strings"

	jwtmgr "github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/jwt"
)

type tokenValidator interface {
	ValidateToken(tokenStr string, expectedAud string) (*jwtmgr.Claims, error)
}

type contextKey string

const ctxClaimsKey contextKey = "claims"

func WithAuth(j tokenValidator, expectedAud string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		tok := strings.TrimPrefix(h, "Bearer ")
		claims, err := j.ValidateToken(tok, expectedAud)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), ctxClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ClaimsFromContext(r *http.Request) *jwtmgr.Claims {
	if v := r.Context().Value(ctxClaimsKey); v != nil {
		if c, ok := v.(*jwtmgr.Claims); ok {
			return c
		}
	}
	return nil
}
