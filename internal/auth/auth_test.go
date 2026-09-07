package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestHashPasswordAndCheckPassword(t *testing.T) {
	const password = "correct horse battery staple"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hash == password {
		t.Fatal("HashPassword() returned the plaintext password")
	}

	matched, err := CheckPassword(password, hash)
	if err != nil {
		t.Fatalf("CheckPassword() error = %v", err)
	}
	if !matched {
		t.Fatal("CheckPassword() = false for the matching password")
	}

	matched, err = CheckPassword("wrong password", hash)
	if err != nil {
		t.Fatalf("CheckPassword() for an incorrect password error = %v", err)
	}
	if matched {
		t.Fatal("CheckPassword() = true for an incorrect password")
	}

	matched, err = CheckPassword(password, "not-a-valid-argon2-hash")
	if err == nil {
		t.Fatal("CheckPassword() error = nil for an invalid hash")
	}
	if matched {
		t.Fatal("CheckPassword() = true for an invalid hash")
	}
}

func TestMakeJWTAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	const secret = "test-secret"

	token, err := MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT() error = %v", err)
	}

	got, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT() error = %v", err)
	}
	if got != userID {
		t.Errorf("ValidateJWT() = %v, want %v", got, userID)
	}
}

func TestValidateJWTRejectsInvalidTokens(t *testing.T) {
	userID := uuid.New()
	const secret = "test-secret"

	token, err := MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT() error = %v", err)
	}

	tests := []struct {
		name  string
		token string
		key   string
	}{
		{name: "malformed token", token: "not-a-jwt", key: secret},
		{name: "wrong signing key", token: token, key: "wrong-secret"},
		{name: "expired token", token: mustMakeJWT(t, userID.String(), secret, -time.Hour)},
		{name: "invalid subject", token: mustMakeJWT(t, "not-a-uuid", secret, time.Hour)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateJWT(tt.token, tt.key)
			if err == nil {
				t.Fatal("ValidateJWT() error = nil, want an error")
			}
			if got != uuid.Nil {
				t.Errorf("ValidateJWT() ID = %v, want uuid.Nil", got)
			}
		})
	}
}

func mustMakeJWT(t *testing.T, subject, secret string, expiresIn time.Duration) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   subject,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
	})
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}
	return signed
}
