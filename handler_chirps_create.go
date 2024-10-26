package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/chonginator/chirpy/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID uuid.UUID 				`json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body string 				`json:"body"`
	UserID uuid.UUID 		`json:"user_id"`
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string 		 `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error decoding parameters", err)
		return
	}

	if params.Body == "" {
		err := fmt.Errorf("body field is empty")
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	err = validateChirp(params.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	cleanedChirp := cleanChirp(params.Body)
	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body: cleanedChirp,
		UserID: params.UserID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
			ID: chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body: chirp.Body,
			UserID: chirp.UserID,
		},
	)
}

func validateChirp(chirp string) error {
	const maxChirpLength = 140
	if len(chirp) > maxChirpLength {
		err := fmt.Errorf("chirp exceeds %d characters", maxChirpLength)
		return err
	}
	return nil
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