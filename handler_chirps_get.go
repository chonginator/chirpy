package main

import (
	"net/http"
	"slices"

	"github.com/google/uuid"
)

const (
	sortAscending  = "asc"
	sortDescending = "desc"
)

func (cfg *apiConfig) handlerChirpsRetrieve(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error getting dbChirps", err)
		return
	}

	authorID := uuid.Nil
	authorIDString := r.URL.Query().Get("author_id")
	if authorIDString != "" {
		authorID, err = uuid.Parse(authorIDString)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid user ID", err)
			return
		}
	}

	sortBy := r.URL.Query().Get("sort")
	if sortBy != sortAscending && sortBy != sortDescending {
		sortBy = sortAscending
	}

	chirpsResponse := []Chirp{}
	for _, chirp := range dbChirps {
		if authorID != uuid.Nil && chirp.UserID != authorID {
			continue
		}

		chirpsResponse = append(chirpsResponse, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	slices.SortFunc(chirpsResponse, func(chirpA, chirpB Chirp) int {
		if sortBy == sortAscending {
			return chirpA.CreatedAt.Compare(chirpB.CreatedAt)
		}
		return -chirpA.CreatedAt.Compare(chirpB.CreatedAt)
	})

	respondWithJSON(w, http.StatusOK, chirpsResponse)
}

func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, r *http.Request) {
	chirpIDString := r.PathValue("chirpID")
	chirpUUID, err := uuid.Parse(chirpIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", err)
		return
	}

	dbChirp, err := cfg.db.GetChirpByID(r.Context(), chirpUUID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Error couldn't get chirp", err)
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	})
}
