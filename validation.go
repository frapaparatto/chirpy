package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func validationHandler(w http.ResponseWriter, r *http.Request) {
	type reqParams struct {
		Body string `json:"body"`
	}

	type validResponse struct {
		Body string `json:"cleaned_body"`
	}

	params := reqParams{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	var cleaned []string
	const subWord = "****"
	for word := range strings.SplitSeq(params.Body, " ") {
		lowered := strings.ToLower(word)
		if lowered == "kerfuffle" || lowered == "sharbert" || lowered == "fornax" {
			word = subWord
		}
		cleaned = append(cleaned, word)
	}

	respString := strings.Join(cleaned, " ")

	respondWithJSON(w, http.StatusOK, validResponse{
		Body: respString,
	})

}
