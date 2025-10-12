// Package auth provides utility functions for authentication validation.
package auth

import (
	"errors"
	"os"
	"strconv"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}

func CheckPasswordHash(pw, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw))
}

func ValidatePasswordPolicy(pw string) error {
	minLength := 12
	if v := os.Getenv("PASSWORD_MIN_LENGTH"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			minLength = n
		}
	}
	if len(pw) < minLength {
		return errors.New("password too short")
	}
	var upper, lower, digit, symbol bool
	for _, r := range pw {
		switch {
		case unicode.IsUpper(r):
			upper = true
		case unicode.IsLower(r):
			lower = true
		case unicode.IsDigit(r):
			digit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			symbol = true
		}
	}
	if !(upper && lower && digit && symbol) {
		return errors.New("password must include upper, lower, digit, and symbol")
	}
	return nil
}
