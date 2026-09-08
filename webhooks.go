package main

import (
	"encoding/json"
	"net/http"

	"github.com/frapaparatto/chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *Config) handleWebhook(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, "Couldn't find API-Key", err)
		return
	}

	if apiKey != cfg.apiKey {
		writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	type Params struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	data := Params{}
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&data); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if data.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
	}

	if err := cfg.db.UpdateUserStatus(r.Context(), data.Data.UserID); err != nil {
		writeErrorResponse(w, http.StatusNotFound, "User not found", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
