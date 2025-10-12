package grpcserver

import (
	"context"
	"testing"
	"time"

	gen "github.com/aezizhu/million-dollar-hunter/services/auth-service/api/gen"
	jwtmgr "github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/jwt"
)

func TestGenerateTokens(t *testing.T) {
	m := jwtmgr.New("iss", "aud", time.Minute, time.Hour, []byte("key"))
	s := New(m)

	resp, err := s.GenerateTokens(context.Background(), &gen.TokenRequest{
		UserId: "u123",
		Email:  "e@example.com",
	})
	if err != nil {
		t.Fatalf("GenerateTokens error: %v", err)
	}
	if resp.GetAccessToken() == "" || resp.GetRefreshToken() == "" || resp.GetExpiresIn() <= 0 {
		t.Fatalf("invalid token pair: %+v", resp)
	}
}

func TestValidateToken_Success(t *testing.T) {
	m := jwtmgr.New("iss", "aud", time.Minute, time.Hour, []byte("key"))
	s := New(m)

	access, _, err := m.GenerateToken("u123", "e@example.com", time.Minute)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}
	resp, err := s.ValidateToken(context.Background(), &gen.ValidateRequest{
		Token:       access,
		ExpectedAud: "aud",
	})
	if err != nil {
		t.Fatalf("ValidateToken error: %v", err)
	}
	if !resp.GetValid() || resp.GetUserId() != "u123" || resp.GetEmail() != "e@example.com" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestValidateToken_WrongAud(t *testing.T) {
	m := jwtmgr.New("iss", "aud", time.Minute, time.Hour, []byte("key"))
	s := New(m)

	access, _, err := m.GenerateToken("u123", "e@example.com", time.Minute)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}
	resp, err := s.ValidateToken(context.Background(), &gen.ValidateRequest{
		Token:       access,
		ExpectedAud: "other-aud",
	})
	if err != nil {
		t.Fatalf("ValidateToken error: %v", err)
	}
	if resp.GetValid() || resp.GetReason() == "" {
		t.Fatalf("expected invalid with reason, got: %+v", resp)
	}
}

func TestValidateToken_Malformed(t *testing.T) {
	m := jwtmgr.New("iss", "aud", time.Minute, time.Hour, []byte("key"))
	s := New(m)
	resp, err := s.ValidateToken(context.Background(), &gen.ValidateRequest{
		Token:       "not.a.jwt",
		ExpectedAud: "aud",
	})
	if err != nil {
		t.Fatalf("ValidateToken error: %v", err)
	}
	if resp.GetValid() || resp.GetReason() == "" {
		t.Fatalf("expected invalid with reason, got: %+v", resp)
	}
}

func TestValidateToken_Expired(t *testing.T) {
	m := jwtmgr.New("iss", "aud", time.Minute, time.Hour, []byte("key"))
	s := New(m)
	access, _, err := m.GenerateToken("u123", "e@example.com", -1*time.Minute)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}
	resp, err := s.ValidateToken(context.Background(), &gen.ValidateRequest{
		Token:       access,
		ExpectedAud: "aud",
	})
	if err != nil {
		t.Fatalf("ValidateToken error: %v", err)
	}
	if resp.GetValid() || resp.GetReason() == "" {
		t.Fatalf("expected invalid expired, got: %+v", resp)
	}
}

func TestValidateToken_WrongIssuer(t *testing.T) {
	m := jwtmgr.New("iss", "aud", time.Minute, time.Hour, []byte("key"))
	s := New(m)
	other := jwtmgr.New("other-iss", "aud", time.Minute, time.Hour, []byte("key"))
	access, _, err := other.GenerateToken("u123", "e@example.com", time.Minute)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}
	resp, err := s.ValidateToken(context.Background(), &gen.ValidateRequest{
		Token:       access,
		ExpectedAud: "aud",
	})
	if err != nil {
		t.Fatalf("ValidateToken error: %v", err)
	}
	if resp.GetValid() || resp.GetReason() == "" {
		t.Fatalf("expected invalid wrong issuer, got: %+v", resp)
	}
}

func TestValidateToken_CancelledContext(t *testing.T) {
	m := jwtmgr.New("iss", "aud", time.Minute, time.Hour, []byte("key"))
	s := New(m)
	access, _, err := m.GenerateToken("u123", "e@example.com", time.Minute)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	resp, err := s.ValidateToken(ctx, &gen.ValidateRequest{
		Token:       access,
		ExpectedAud: "aud",
	})
	if err != nil {
		t.Fatalf("unexpected error on canceled context: %v", err)
	}
	if !resp.GetValid() || resp.GetUserId() != "u123" || resp.GetEmail() != "e@example.com" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}
