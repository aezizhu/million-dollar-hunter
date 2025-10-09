package auth

import "testing"

func TestValidatePasswordPolicy(t *testing.T) {
	if err := ValidatePasswordPolicy("Short1!"); err == nil {
		t.Fatalf("expected error for short password")
	}
	if err := ValidatePasswordPolicy("longbutnosymbol1A"); err == nil {
		t.Fatalf("expected error for missing symbol")
	}
	if err := ValidatePasswordPolicy("ValidPass12!"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHashAndCheckPassword(t *testing.T) {
	pw := "ValidPass12!"
	hash, err := HashPassword(pw)
	if err != nil {
		t.Fatalf("hash error: %v", err)
	}
	if err := CheckPasswordHash(pw, hash); err != nil {
		t.Fatalf("check error: %v", err)
	}
	if err := CheckPasswordHash("wrong", hash); err == nil {
		t.Fatalf("expected mismatch error")
	}
}
