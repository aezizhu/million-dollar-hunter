package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	mw "github.com/aezizhu/million-dollar-hunter/api-gateway/internal/middleware"
	rl "github.com/aezizhu/million-dollar-hunter/api-gateway/internal/ratelimit"
)

func setupRedis(t *testing.T, ctx context.Context) (testcontainers.Container, *redis.Client) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(30 * time.Second),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{ContainerRequest: req, Started: true})
	require.NoError(t, err)
	host, err := c.Host(ctx)
	require.NoError(t, err)
	port, err := c.MappedPort(ctx, "6379")
	require.NoError(t, err)

	rdb := redis.NewClient(&redis.Options{Addr: host + ":" + port.Port()})
	return c, rdb
}

func TestRedisTokenBucket_AllowAndBlock(t *testing.T) {
	if testing.Short() {
		t.Skip("short")
	}
	ctx := context.Background()
	c, rdb := setupRedis(t, ctx)
	defer testcontainers.TerminateContainer(c)
	defer rdb.Close()

	bucket := rl.NewRedisTokenBucket(rdb, 10, 10, time.Second, "it-key")
	for i := 0; i < 10; i++ {
		allowed, _, remaining, reset, retry := bucket.Allow(ctx, "k")
		assert.True(t, allowed)
		assert.GreaterOrEqual(t, remaining, 0)
		assert.True(t, reset.After(time.Now()))
		assert.Equal(t, time.Duration(0), retry)
	}
	allowed, _, remaining, _, retry := bucket.Allow(ctx, "k")
	assert.False(t, allowed)
	assert.Equal(t, 0, remaining)
	assert.Greater(t, retry, time.Duration(0))
}

type limiterAdapter struct{ b *rl.RedisTokenBucket }

func (l limiterAdapter) Allow(key string) (bool, int, int, time.Time, time.Duration) {
	return l.b.Allow(context.Background(), key)
}

func TestRateLimitMiddleware_WithRedis(t *testing.T) {
	if testing.Short() {
		t.Skip("short")
	}
	ctx := context.Background()
	c, rdb := setupRedis(t, ctx)
	defer testcontainers.TerminateContainer(c)
	defer rdb.Close()

	bucket := rl.NewRedisTokenBucket(rdb, 2, 2, 1*time.Second, "it-key")
	time.Sleep(50 * time.Millisecond)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(mw.RateLimit(limiterAdapter{bucket}))
	r.GET("/api/v1/auth/login", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/login", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if i < 2 {
			assert.Equal(t, 200, w.Code)
		} else {
			assert.Equal(t, 429, w.Code)
			limitHdr := w.Header().Get("X-RateLimit-Limit")
			if limitHdr == "" {
				limitHdr = w.Header().Get("RateLimit-Limit")
			}
			assert.NotEmpty(t, limitHdr, "expected ratelimit header")
			assert.NotEmpty(t, w.Header().Get("Retry-After"))
		}
	}
}
