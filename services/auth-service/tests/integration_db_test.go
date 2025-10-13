package tests

import (
	"context"
	"database/sql"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/bcrypt"

	jwtmgr "github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/jwt"
	"github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/store"
)

func TestDatabaseConnectivity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("auth_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second)),
	)
	require.NoError(t, err)
	defer func() {
		if termErr := testcontainers.TerminateContainer(pgContainer); termErr != nil {
			t.Logf("failed to terminate container: %s", termErr)
		}
	}()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err, "Failed to connect to test database")
	defer pool.Close()

	err = pool.Ping(ctx)
	require.NoError(t, err, "Failed to ping database")

	err = runMigrations(ctx, pool)
	require.NoError(t, err, "Failed to run migrations")

	var tableCount int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name IN ('users', 'refresh_tokens', 'auth_audit')
	`).Scan(&tableCount)
	require.NoError(t, err)
	assert.Equal(t, 3, tableCount, "Expected 3 tables to be created")

	var uuidExtExists bool
	err = pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'uuid-ossp')
	`).Scan(&uuidExtExists)
	require.NoError(t, err)
	assert.True(t, uuidExtExists, "uuid-ossp extension should be installed")
}

func TestJWTFlowEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pgContainer, pool := setupDatabase(t, ctx)
	defer testcontainers.TerminateContainer(pgContainer)
	defer pool.Close()

	s := store.New(pool)
	jwtManager := jwtmgr.New("test-issuer", "test-audience", 15*time.Minute, 168*time.Hour, []byte("test-secret-key"))

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("TestPassword123!"), bcrypt.DefaultCost)
	require.NoError(t, err)

	var userID string
	email := "test@example.com"
	err = pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id
	`, email, string(passwordHash)).Scan(&userID)
	require.NoError(t, err)

	accessToken, refreshToken, expiresAt, err := jwtManager.GeneratePair(userID, email)
	require.NoError(t, err, "Failed to generate token pair")
	assert.NotEmpty(t, accessToken, "Access token should not be empty")
	assert.NotEmpty(t, refreshToken, "Refresh token should not be empty")
	assert.True(t, expiresAt.After(time.Now()), "Expiration should be in the future")

	err = s.StoreRefreshToken(ctx, userID, refreshToken, expiresAt)
	require.NoError(t, err, "Failed to store refresh token")

	claims, err := jwtManager.ValidateToken(accessToken, "test-audience")
	require.NoError(t, err, "Failed to validate access token")
	assert.Equal(t, userID, claims.UserID, "UserID should match")
	assert.Equal(t, email, claims.Email, "Email should match")
	assert.Equal(t, "test-issuer", claims.Issuer, "Issuer should match")
	assert.Equal(t, "test-audience", claims.Audience[0], "Audience should match")

	var storedToken string
	err = pool.QueryRow(ctx, `
		SELECT token FROM refresh_tokens 
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID).Scan(&storedToken)
	require.NoError(t, err)
	assert.Equal(t, refreshToken, storedToken, "Stored refresh token should match")
}

func TestRefreshTokenRotation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pgContainer, pool := setupDatabase(t, ctx)
	defer testcontainers.TerminateContainer(pgContainer)
	defer pool.Close()

	s := store.New(pool)
	jwtManager := jwtmgr.New("test-issuer", "test-audience", 15*time.Minute, 168*time.Hour, []byte("test-secret-key"))

	var userID string
	email := "rotate@example.com"
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("TestPassword123!"), bcrypt.DefaultCost)
	err := pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id
	`, email, string(passwordHash)).Scan(&userID)
	require.NoError(t, err)

	_, oldRefreshToken, expiresAt, err := jwtManager.GeneratePair(userID, email)
	require.NoError(t, err)

	err = s.StoreRefreshToken(ctx, userID, oldRefreshToken, expiresAt)
	require.NoError(t, err)

	var revokedAt sql.NullTime
	err = pool.QueryRow(ctx, `
		SELECT revoked_at FROM refresh_tokens WHERE token = $1
	`, oldRefreshToken).Scan(&revokedAt)
	require.NoError(t, err)
	assert.False(t, revokedAt.Valid, "Old token should not be revoked initially")

	err = s.RevokeRefreshToken(ctx, oldRefreshToken)
	require.NoError(t, err)

	newAccessToken, newRefreshToken, newExpiresAt, err := jwtManager.GeneratePair(userID, email)
	require.NoError(t, err)

	err = s.StoreRefreshToken(ctx, userID, newRefreshToken, newExpiresAt)
	require.NoError(t, err)

	err = pool.QueryRow(ctx, `
		SELECT revoked_at FROM refresh_tokens WHERE token = $1
	`, oldRefreshToken).Scan(&revokedAt)
	require.NoError(t, err)
	assert.True(t, revokedAt.Valid, "Old token should be revoked")

	err = pool.QueryRow(ctx, `
		SELECT revoked_at FROM refresh_tokens WHERE token = $1
	`, newRefreshToken).Scan(&revokedAt)
	require.NoError(t, err)
	assert.False(t, revokedAt.Valid, "New token should not be revoked")

	claims, err := jwtManager.ValidateToken(newAccessToken, "test-audience")
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)

	isValid, _, err := s.ValidateRefreshToken(ctx, oldRefreshToken)
	require.NoError(t, err)
	assert.False(t, isValid, "Revoked token should not be valid")
}

func TestLoginLockout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pgContainer, pool := setupDatabase(t, ctx)
	defer testcontainers.TerminateContainer(pgContainer)
	defer pool.Close()

	s := store.New(pool)

	email := "lockout@example.com"
	var userID string
	correctPassword := "CorrectPassword123!"
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)

	err := pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id
	`, email, string(passwordHash)).Scan(&userID)
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		err = s.RecordLoginAttempt(ctx, userID, "login_failed", "127.0.0.1", "test-agent")
		require.NoError(t, err)
	}

	isLocked, err := s.IsUserLockedOut(ctx, userID, 3, 15*time.Minute)
	require.NoError(t, err)
	assert.True(t, isLocked, "User should be locked out after 3 failed attempts")

	var count int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM auth_audit 
		WHERE user_id = $1 AND event_type = 'login_failed'
		AND timestamp > NOW() - INTERVAL '15 minutes'
	`, userID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 3, count, "Should have 3 failed login records")
}

func TestTokenExpiration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	jwtManager := jwtmgr.New("test-issuer", "test-audience", 1*time.Millisecond, 1*time.Millisecond, []byte("test-secret-key"))

	var userID string
	email := "expire@example.com"

	accessToken, _, _, err := jwtManager.GeneratePair(userID, email)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	_, err = jwtManager.ValidateToken(accessToken, "test-audience")
	assert.Error(t, err, "Expired token should fail validation")
	assert.Contains(t, err.Error(), "expired", "Error should mention expiration")
}

func TestConcurrentTokenGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pgContainer, pool := setupDatabase(t, ctx)
	defer testcontainers.TerminateContainer(pgContainer)
	defer pool.Close()

	s := store.New(pool)
	jwtManager := jwtmgr.New("test-issuer", "test-audience", 15*time.Minute, 168*time.Hour, []byte("test-secret-key"))

	var userID string
	email := "concurrent@example.com"
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("TestPassword123!"), bcrypt.DefaultCost)
	err := pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id
	`, email, string(passwordHash)).Scan(&userID)
	require.NoError(t, err)

	concurrency := 10
	done := make(chan bool, concurrency)
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			_, refreshToken, expiresAt, genErr := jwtManager.GeneratePair(userID, email)
			if genErr != nil {
				errors <- genErr
				done <- false
				return
			}

			storeErr := s.StoreRefreshToken(ctx, userID, refreshToken, expiresAt)
			if storeErr != nil {
				errors <- storeErr
				done <- false
				return
			}

			done <- true
		}()
	}

	successCount := 0
	for i := 0; i < concurrency; i++ {
		if <-done {
			successCount++
		}
	}
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent operation failed: %v", err)
	}

	assert.Equal(t, concurrency, successCount, "All concurrent operations should succeed")

	var tokenCount int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM refresh_tokens WHERE user_id = $1
	`, userID).Scan(&tokenCount)
	require.NoError(t, err)
	assert.Equal(t, concurrency, tokenCount, "All refresh tokens should be stored")
}
func TestConcurrentRefreshRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pgContainer, pool := setupDatabase(t, ctx)
	defer testcontainers.TerminateContainer(pgContainer)
	defer pool.Close()

	s := store.New(pool)
	jwtManager := jwtmgr.New("test-issuer", "test-audience", 15*time.Minute, 168*time.Hour, []byte("test-secret-key"))

	var userID string
	email := "refresh-race@example.com"
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("SomePass123!"), bcrypt.DefaultCost)
	err := pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id
	`, email, string(passwordHash)).Scan(&userID)
	require.NoError(t, err)

	_, refreshToken, expiresAt, err := jwtManager.GeneratePair(userID, email)
	require.NoError(t, err)
	err = s.StoreRefreshToken(ctx, userID, refreshToken, expiresAt)
	require.NoError(t, err)

	concurrency := 10
	var success int32
	wg := sync.WaitGroup{}
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			if err := s.RevokeRefreshToken(ctx, refreshToken); err == nil {
				atomic.AddInt32(&success, 1)
			}
		}()
	}
	wg.Wait()

	assert.Equal(t, int32(1), success, "only one concurrent refresh should succeed")

	var revokedAt sql.NullTime
	err = pool.QueryRow(ctx, `
		SELECT revoked_at FROM refresh_tokens WHERE token = $1
	`, refreshToken).Scan(&revokedAt)
	require.NoError(t, err)
	assert.True(t, revokedAt.Valid, "refresh token should be revoked")
}

func TestLockoutBoundaryConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	pgContainer, pool := setupDatabase(t, ctx)
	defer testcontainers.TerminateContainer(pgContainer)
	defer pool.Close()

	s := store.New(pool)

	var userID string
	email := "lock-boundary@example.com"
	pw, _ := bcrypt.GenerateFromPassword([]byte("Password1!"), bcrypt.DefaultCost)
	err := pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id
	`, email, string(pw)).Scan(&userID)
	require.NoError(t, err)

	for i := 0; i < 2; i++ {
		require.NoError(t, s.RecordLoginAttempt(ctx, userID, "login_failed", "127.0.0.1", "test-agent"))
	}
	locked, err := s.IsUserLockedOut(ctx, userID, 3, 15*time.Minute)
	require.NoError(t, err)
	assert.False(t, locked, "N-1 failures should not lock")

	require.NoError(t, s.RecordLoginAttempt(ctx, userID, "login_failed", "127.0.0.1", "test-agent"))
	locked, err = s.IsUserLockedOut(ctx, userID, 3, 15*time.Minute)
	require.NoError(t, err)
	assert.True(t, locked, "N failures should lock")

	require.NoError(t, s.RecordLoginAttempt(ctx, userID, "login_failed", "127.0.0.1", "test-agent"))
	locked, err = s.IsUserLockedOut(ctx, userID, 3, 15*time.Minute)
	require.NoError(t, err)
	assert.True(t, locked, "N+1 failures remain locked")
}

func TestLockoutWindowExpiration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	pgContainer, pool := setupDatabase(t, ctx)
	defer testcontainers.TerminateContainer(pgContainer)
	defer pool.Close()

	s := store.New(pool)

	var userID string
	email := "lock-window@example.com"
	pw, _ := bcrypt.GenerateFromPassword([]byte("Password1!"), bcrypt.DefaultCost)
	err := pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id
	`, email, string(pw)).Scan(&userID)
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		require.NoError(t, s.RecordLoginAttempt(ctx, userID, "login_failed", "127.0.0.1", "test-agent"))
	}
	locked, err := s.IsUserLockedOut(ctx, userID, 3, 15*time.Minute)
	require.NoError(t, err)
	assert.True(t, locked, "should be locked after N failures")

	_, err = pool.Exec(ctx, `
		UPDATE auth_audit
		SET timestamp = NOW() - INTERVAL '16 minutes'
		WHERE user_id = $1 AND event_type = 'login_failed'
	`, userID)
	require.NoError(t, err)

	locked, err = s.IsUserLockedOut(ctx, userID, 3, 15*time.Minute)
	require.NoError(t, err)
	assert.False(t, locked, "lockout should expire after window")
}

func TestLockoutPerUserIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	pgContainer, pool := setupDatabase(t, ctx)
	defer testcontainers.TerminateContainer(pgContainer)
	defer pool.Close()

	s := store.New(pool)

	var userA, userB string
	pw, _ := bcrypt.GenerateFromPassword([]byte("Password1!"), bcrypt.DefaultCost)

	err := pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id
	`, "userA@example.com", string(pw)).Scan(&userA)
	require.NoError(t, err)

	err = pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id
	`, "userB@example.com", string(pw)).Scan(&userB)
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		require.NoError(t, s.RecordLoginAttempt(ctx, userA, "login_failed", "127.0.0.1", "test-agent"))
	}

	lockedA, err := s.IsUserLockedOut(ctx, userA, 3, 15*time.Minute)
	require.NoError(t, err)
	assert.True(t, lockedA, "user A should be locked")

	lockedB, err := s.IsUserLockedOut(ctx, userB, 3, 15*time.Minute)
	require.NoError(t, err)
	assert.False(t, lockedB, "user B should not be locked")
}


func setupDatabase(t *testing.T, ctx context.Context) (testcontainers.Container, *pgxpool.Pool) {
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("auth_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second)),
	)
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	err = runMigrations(ctx, pool)
	require.NoError(t, err)

	return pgContainer, pool
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrations := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
		
		CREATE OR REPLACE FUNCTION set_updated_at()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
		
		DROP TRIGGER IF EXISTS trg_users_set_updated_at ON users;
		CREATE TRIGGER trg_users_set_updated_at
		BEFORE UPDATE ON users
		FOR EACH ROW
		EXECUTE FUNCTION set_updated_at();`,

		`CREATE TABLE IF NOT EXISTS refresh_tokens (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token TEXT NOT NULL UNIQUE,
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			revoked_at TIMESTAMPTZ
		);
		CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON refresh_tokens(user_id);
		CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires ON refresh_tokens(expires_at);`,

		`CREATE TABLE IF NOT EXISTS auth_audit (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			event_type TEXT NOT NULL,
			ip_address TEXT,
			user_agent TEXT,
			timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_auth_audit_user ON auth_audit(user_id);
		CREATE INDEX IF NOT EXISTS idx_auth_audit_timestamp ON auth_audit(timestamp);`,
	}

	for _, migration := range migrations {
		_, err := pool.Exec(ctx, migration)
		if err != nil {
			return err
		}
	}

	return nil
}
