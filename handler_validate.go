package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func handlerChirpsValidate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding parameters", err)
		return
	}

	if params.Body == "" {
		respondWithError(w, http.StatusBadRequest, "Body field is required", nil)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Chirp exceeds %d characters", maxChirpLength), nil)
		return
	}

	cleanedChirp := cleanChirp(params.Body)
	respondWithJSON(w, http.StatusOK, returnVals{ CleanedBody: cleanedChirp })
}

func cleanChirp(chirp string) string {
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert": {},
		"fornax": {},
	}

	words := strings.Split(chirp, " ")
	for i, char := range(words) {
		if _, ok := badWords[strings.ToLower(char)]; ok {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")
}