package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/frapaparatto/chirpy/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *Config) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	type ChirpData struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	chirpReq := ChirpData{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&chirpReq); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	if chirpReq.UserID == uuid.Nil {
		writeErrorResponse(w, http.StatusBadRequest, "user_id is required", nil)
		return
	}

	cleanedBody, err := cleanChirpBody(chirpReq.Body)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Chirp too long", err)
		return
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: chirpReq.UserID,
	})

	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to create chirp", err)
		return
	}

	writeJSONResponse(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func cleanChirpBody(body string) (string, error) {
	var cleaned []string
	const subWord = "****"

	if len(body) > 140 {
		return "", errors.New("Too long chirp")
	}

	for word := range strings.SplitSeq(body, " ") {
		lowered := strings.ToLower(word)
		if lowered == "kerfuffle" || lowered == "sharbert" || lowered == "fornax" {
			word = subWord
		}
		cleaned = append(cleaned, word)
	}

	respString := strings.Join(cleaned, " ")

	return respString, nil
}

func (cfg *Config) handleListChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.ListChirps(r.Context())
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Could not list chirps", err)
		return
	}

	responseChirps := make([]Chirp, len(chirps))
	for i, chirp := range chirps {
		responseChirps[i] = Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
	}

	writeJSONResponse(w, http.StatusOK, responseChirps)
}

func (cfg *Config) handleGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	parsedID, err := uuid.Parse(chirpID)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid chirp ID", err)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), parsedID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeErrorResponse(w, http.StatusNotFound, "Chirp not found", nil)
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Could not get chirp", err)
		return
	}

	writeJSONResponse(w, http.StatusOK, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})

}
