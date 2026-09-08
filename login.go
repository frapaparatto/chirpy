package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/frapaparatto/chirpy/internal/auth"
	"github.com/frapaparatto/chirpy/internal/database"
)

const refreshTokenExpiration = 60 * 24 * time.Hour

func (cfg *Config) handleLogin(w http.ResponseWriter, r *http.Request) {
	type LoginData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	usr := LoginData{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&usr); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	user, err := cfg.db.GetByEmail(r.Context(), usr.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeErrorResponse(w, http.StatusUnauthorized, "Invalid credentials", nil)
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Could not log in", err)
		return
	}

	match, err := auth.CheckPassword(usr.Password, user.HashedPassword)

	if err != nil || !match {
		writeErrorResponse(w, http.StatusUnauthorized, "Invalid credentials", err)
		return
	}

	expiresIn := time.Hour
	token, err := auth.MakeJWT(user.ID, cfg.secretKey, expiresIn)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Could not create token", err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to generate refresh token", err)
		return
	}
	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(refreshTokenExpiration),
	})

	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to create refresh token", err)
		return
	}

	writeJSONResponse(w, http.StatusOK, User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		JWTToken:     token,
		RefreshToken: refreshToken,
		IsChirpyRed:  user.IsChirpyRed,
	})

}
