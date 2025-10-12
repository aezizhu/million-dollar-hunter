package store

import (
	"context"
	"time"
)

type AuditEvent struct {
	ID        int64
	UserID    *string
	Event     string
	IP        *string
	UserAgent *string
	CreatedAt time.Time
}

type AuditStore interface {
	Log(ctx context.Context, userID *string, event string, ip *string, ua *string) error
	CountRecentLoginFailures(ctx context.Context, userID *string, window time.Duration) (int, error)
}

func (s *PGStore) Log(ctx context.Context, userID *string, event string, ip *string, ua *string) error {
	const q = `INSERT INTO auth_audit (user_id, event_type, ip_address, user_agent) VALUES ($1,$2,$3,$4)`
	_, err := s.Pool.Exec(ctx, q, userID, event, ip, ua)
	return err
}

func (s *PGStore) RecordLoginAttempt(ctx context.Context, userID, eventType, ipAddress, userAgent string) error {
	uid := &userID
	ip := &ipAddress
	ua := &userAgent
	return s.Log(ctx, uid, eventType, ip, ua)
}

func (s *PGStore) IsUserLockedOut(ctx context.Context, userID string, maxAttempts int, windowDuration time.Duration) (bool, error) {
	const q = `SELECT COUNT(*) FROM auth_audit 
	           WHERE user_id = $1 AND event_type = 'login_failed'
	           AND timestamp > NOW() - $2::interval`

	var count int
	interval := windowDuration.String()
	err := s.Pool.QueryRow(ctx, q, userID, interval).Scan(&count)
	if err != nil {
		return false, err
	}

	return count >= maxAttempts, nil
}

func (s *PGStore) CountRecentLoginFailures(ctx context.Context, userID *string, window time.Duration) (int, error) {
	const q = `SELECT COUNT(*) FROM auth_audit WHERE event='login_failure' AND created_at > NOW() - $1::interval AND (user_id IS NOT DISTINCT FROM $2)`
	var n int
	interval := window.String()
	err := s.Pool.QueryRow(ctx, q, interval, userID).Scan(&n)
	return n, err
}
