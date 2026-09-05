package main

import (
	"fmt"
	"net/http"
)

func (cfg *Config) resetHandler(w http.ResponseWriter, r *http.Request) {

	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Forbidden operation", nil)
		return
	}

	if err := cfg.db.ResetUsers(r.Context()); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8") // normal header
	t := fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())
	w.Write([]byte(t))

}
