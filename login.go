package main

import (
	"encoding/json"
	"net/http"

	"github.com/frapaparatto/chirpy/internal/auth"
)

func (cfg *Config) handleUserLogin(w http.ResponseWriter, r *http.Request) {
	type LoginData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	usr := LoginData{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&usr); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	user, err := cfg.db.GetByEmail(r.Context(), usr.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "User not found", err)
		return
	}

	match, err := auth.CheckPasswordHash(usr.Password, user.HashedPassword)

	if err != nil || !match {
		respondWithError(w, http.StatusUnauthorized, "User unauthorized", err)
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})

}
