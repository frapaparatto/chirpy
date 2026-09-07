package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	return hash, err
}

func CheckPassword(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	return match, err
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	})

	return token.SignedString([]byte(tokenSecret))

}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{},
		func(token *jwt.Token) (any, error) {
			return []byte(tokenSecret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil {
		return uuid.Nil, err
	}

	id, err := token.Claims.GetSubject()

	if err != nil {
		return uuid.Nil, err
	}

	return uuid.Parse(id)

}

func GetBearerToken(headers http.Header) (string, error) {
	const prefix = "Bearer "

	token := headers.Get("Authorization")
	if token == "" {
		return "", errors.New("No token were provided")
	}

	if !strings.HasPrefix(token, prefix) {
		return "", errors.New("Authorization header must use Bearer scheme")
	}

	return strings.TrimSpace(strings.TrimPrefix(token, prefix)), nil
}

func MakeRefreshToken() string {
	key := make([]byte, 32)
	rand.Read(key)

	return hex.EncodeToString(key)
}
