package main

import (
	"fmt"
	"net/http"
)

func (cfg *Config) handleReset(w http.ResponseWriter, r *http.Request) {

	if cfg.platform != "dev" {
		writeErrorResponse(w, http.StatusForbidden, "Forbidden", nil)
		return
	}

	if err := cfg.db.ResetUsers(r.Context()); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Could not reset", err)
		return
	}

	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8") // normal header
	t := fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())
	w.Write([]byte(t))

}
