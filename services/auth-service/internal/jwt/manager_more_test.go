package jwtmgr

import (
	"testing"
	"time"
)

func TestGeneratePair_CoversBranches(t *testing.T) {
	m := New("iss", "aud", time.Minute, time.Hour, []byte("k"))
	access, refresh, exp, err := m.GeneratePair("u1", "u@example.com")
	if err != nil {
		t.Fatalf("GeneratePair error: %v", err)
	}
	if access == "" || refresh == "" {
		t.Fatalf("tokens empty")
	}
	if exp.Before(time.Now()) {
		t.Fatalf("exp in past")
	}
	access2, refresh2, _, err := m.GeneratePair("u1", "u@example.com")
	if err != nil {
		t.Fatalf("GeneratePair2 error: %v", err)
	}
	if access2 == access || refresh2 == refresh {
		t.Fatalf("tokens should differ across calls")
	}
}
