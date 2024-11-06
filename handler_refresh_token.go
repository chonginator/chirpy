package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/chonginator/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find refresh token", err)
		return
	}

	dbRefreshToken, err := cfg.db.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error getting refresh token from database", err)
		return
	}

	if dbRefreshToken.RevokedAt.Valid {
		err := errors.New("refresh token has been revoked")
		respondWithError(w, http.StatusUnauthorized, err.Error(), err)
		return
	}

	if dbRefreshToken.ExpiresAt.Before(time.Now()) {
		err := errors.New("refresh token is expired")
		respondWithError(w, http.StatusUnauthorized, err.Error(), err)
		return
	}

	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), dbRefreshToken.UserID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get user from refresh token", err)
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error making access token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Token: accessToken,
	})
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find refresh token", err)
		return
	}

	dbRefreshToken, err := cfg.db.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error getting refresh token from database", err)
		return
	}

	if dbRefreshToken.ExpiresAt.Before(time.Now()) {
		err := errors.New("refresh token has already expired")
		respondWithError(w, http.StatusUnauthorized, err.Error(), err)
		return
	}

	if dbRefreshToken.RevokedAt.Valid {
		err := errors.New("refresh token has already been revoked")
		respondWithError(w, http.StatusUnauthorized, err.Error(), err)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), dbRefreshToken.Token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't revoke refresh token", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}