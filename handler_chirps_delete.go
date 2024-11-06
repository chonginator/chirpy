package main

import (
	"errors"
	"net/http"

	"github.com/chonginator/chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerChirpsDelete(w http.ResponseWriter, r *http.Request) {
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(accessToken, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	userIDString := r.PathValue("userID")
	userIDFromPath, err := uuid.Parse(userIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	if userID != userIDFromPath {
		err := errors.New("unauthorized action")
		respondWithError(w, http.StatusForbidden, err.Error(), err)
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Error deleting chirp", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
