package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/frapaparatto/chirpy/internal/auth"
)

func (cfg *Config) handleLogin(w http.ResponseWriter, r *http.Request) {
	type LoginData struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		ExpiresIn *int   `json:"expires_in_seconds,omitempty"`
	}

	usr := LoginData{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&usr); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	expiresIn := time.Hour
	if usr.ExpiresIn != nil {
		expiresIn = time.Duration(*usr.ExpiresIn) * time.Second
		if expiresIn <= 0 || expiresIn > time.Hour {
			expiresIn = time.Hour
		}
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

	token, err := auth.MakeJWT(user.ID, cfg.secretKey, expiresIn)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Could not create token", err)
		return
	}

	writeJSONResponse(w, http.StatusOK, User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     token,
	})

}
