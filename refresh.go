package main

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/frapaparatto/chirpy/internal/auth"
)

func (cfg *Config) handleRefresh(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Couldn't find token", err)
		return
	}

	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), token)
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, "Couldn't get user for refresh token", err)
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.secretKey, time.Hour)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Couldn't validate token", err)
		return
	}

	writeJSONResponse(w, http.StatusOK, response{
		Token: accessToken,
	})

}

func (cfg *Config) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Couldn't find token", err)
		return
	}

	_, err = cfg.db.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeErrorResponse(w, http.StatusUnauthorized, "Couldn't revoke session", err)
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Couldn't revoke session", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
