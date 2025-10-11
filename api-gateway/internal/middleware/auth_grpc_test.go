package middleware

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/config"
	gen "github.com/aezizhu/million-dollar-hunter/services/auth-service/api/gen"
)

type fakeAuthServer struct {
	gen.UnimplementedAuthServiceServer
	valid bool
	delay time.Duration
}

func (f *fakeAuthServer) ValidateToken(ctx context.Context, req *gen.ValidateRequest) (*gen.ValidateResponse, error) {
	if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if f.valid {
		return &gen.ValidateResponse{Valid: true, UserId: "u1", Email: "e@x.com"}, nil
	}
	return &gen.ValidateResponse{Valid: false, Reason: "invalid"}, nil
}

func dialBufConn(lis *bufconn.Listener) (*grpc.ClientConn, error) {
	return grpc.DialContext(context.Background(), "bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return lis.Dial()
		}),
	)
}

func startBufServer(s gen.AuthServiceServer) (*bufconn.Listener, *grpc.Server) {
	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	gen.RegisterAuthServiceServer(srv, s)
	go func() { _ = srv.Serve(lis) }()
	return lis, srv
}

func TestAuth_GRPCMode_Success(t *testing.T) {
	lis, srv := startBufServer(&fakeAuthServer{valid: true})
	defer srv.Stop()

	conn, err := dialBufConn(lis)
	if err != nil {
		t.Fatalf("dial err: %v", err)
	}
	defer conn.Close()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := config.Config{
		AuthValidateMode:  "grpc",
		JWTAudience:       "aud",
		AuthGRPCTimeoutMs: 200,
	}
	r.Use(Auth(cfg, conn))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer token")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAuth_GRPCMode_InvalidToken(t *testing.T) {
	lis, srv := startBufServer(&fakeAuthServer{valid: false})
	defer srv.Stop()

	conn, err := dialBufConn(lis)
	if err != nil {
		t.Fatalf("dial err: %v", err)
	}
	defer conn.Close()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := config.Config{
		AuthValidateMode:  "grpc",
		JWTAudience:       "aud",
		AuthGRPCTimeoutMs: 200,
	}
	r.Use(Auth(cfg, conn))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer token")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuth_GRPCMode_Timeout(t *testing.T) {
	lis, srv := startBufServer(&fakeAuthServer{valid: true, delay: 50 * time.Millisecond})
	defer srv.Stop()

	conn, err := dialBufConn(lis)
	if err != nil {
		t.Fatalf("dial err: %v", err)
	}
	defer conn.Close()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := config.Config{
		AuthValidateMode:  "grpc",
		JWTAudience:       "aud",
		AuthGRPCTimeoutMs: 10,
	}
	r.Use(Auth(cfg, conn))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer token")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on timeout, got %d", w.Code)
	}
}

func TestAuth_GRPCMode_NilConn_FallsBackToLocal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := config.Config{
		AuthValidateMode: "grpc",
		JWTSecret:        "devsecret",
	}
	r.Use(Auth(cfg, nil))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer invalid.jwt.token")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid local jwt, got %d", w.Code)
	}
}
func TestAuth_GRPCMode_FallbackToLocal_MVPGate_Succeeds(t *testing.T) {
	lis, srv := startBufServer(&fakeAuthServer{valid: true, delay: 50 * time.Millisecond})
	defer srv.Stop()

	conn, err := dialBufConn(lis)
	if err != nil {
		t.Fatalf("dial err: %v", err)
	}
	defer conn.Close()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := config.Config{
		AuthValidateMode:        "grpc",
		JWTAudience:             "aud",
		AuthGRPCTimeoutMs:       10,
		AuthGRPCFallbackToLocal: true,
		AuthMode:                "mvp-gate",
	}
	r.Use(Auth(cfg, conn))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer 123456789abcdef")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with fallback to local mvp-gate, got %d", w.Code)
	}
}


package middleware

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/aezizhu/million-dollar-hunter/api-gateway/internal/config"
	gen "github.com/aezizhu/million-dollar-hunter/services/auth-service/api/gen"
)

type fakeAuthServer struct {
	gen.UnimplementedAuthServiceServer
	valid bool
}

func (f *fakeAuthServer) ValidateToken(ctx context.Context, req *gen.ValidateRequest) (*gen.ValidateResponse, error) {
	if f.valid {
		return &gen.ValidateResponse{Valid: true, UserId: "u1", Email: "e@x.com"}, nil
	}
	return &gen.ValidateResponse{Valid: false, Reason: "invalid"}, nil
}

func dialBufConn(lis *bufconn.Listener) (*grpc.ClientConn, error) {
	return grpc.DialContext(context.Background(), "bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return lis.Dial()
		}),
	)
}

func startBufServer(s gen.AuthServiceServer) (*bufconn.Listener, *grpc.Server) {
	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	gen.RegisterAuthServiceServer(srv, s)
	go func() { _ = srv.Serve(lis) }()
	return lis, srv
}

func TestAuth_GRPCMode_Success(t *testing.T) {
	lis, srv := startBufServer(&fakeAuthServer{valid: true})
	defer srv.Stop()

	conn, err := dialBufConn(lis)
	if err != nil {
		t.Fatalf("dial err: %v", err)
	}
	defer conn.Close()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := config.Config{
		AuthValidateMode: "grpc",
		JWTAudience:      "aud",
	}
	r.Use(Auth(cfg, conn))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer token")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAuth_GRPCMode_InvalidToken(t *testing.T) {
	lis, srv := startBufServer(&fakeAuthServer{valid: false})
	defer srv.Stop()

	conn, err := dialBufConn(lis)
	if err != nil {
		t.Fatalf("dial err: %v", err)
	}
	defer conn.Close()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := config.Config{
		AuthValidateMode: "grpc",
		JWTAudience:      "aud",
	}
	r.Use(Auth(cfg, conn))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer token")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
