package auth

import "testing"

func TestValidatePasswordPolicy_AllCases(t *testing.T) {
	tests := []struct{
		name string
		pw   string
		ok   bool
	}{
		{"too_short", "Aa1!", false},
		{"no_upper", "validpass12!", false},
		{"no_lower", "VALIDPASS12!", false},
		{"no_digit", "ValidPass!!", false},
		{"no_symbol", "ValidPass12", false},
		{"valid", "ValidPass12!", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordPolicy(tt.pw)
			if tt.ok && err != nil {
				t.Fatalf("expected ok, got err: %v", err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("expected error")
			}
		})
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
