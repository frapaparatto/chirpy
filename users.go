package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/frapaparatto/chirpy/internal/auth"
	"github.com/frapaparatto/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// That is needed for marshaling the json for the response
type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *Config) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	type UserData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	usr := UserData{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&usr); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	hashed_password, err := auth.HashPassword(usr.Password)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to create user", err)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{Email: usr.Email, HashedPassword: hashed_password})
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeErrorResponse(w, http.StatusConflict, "Email already exists", nil)
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to create user", err)
		return
	}

	writeJSONResponse(w, http.StatusCreated, User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})

}
