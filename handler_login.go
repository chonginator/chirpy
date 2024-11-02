package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/chonginator/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string 				`json:"password"`
		Email string 						`json:"email"`
		ExpiresInSeconds string `json:"expires_in_seconds"`
	}
	type response struct {
		User
		Token string						`json:"token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding parameters", err)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	expiresInSecondsDuration := time.Hour
	if params.ExpiresInSeconds != "" {
		expiresInSeconds, err := strconv.Atoi(params.ExpiresInSeconds)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Error parsing expires_in_seconds request body field", err)
			return
		}

		if expiresInSecondsDuration <= time.Hour {
			expiresInSecondsDuration = time.Duration(expiresInSeconds) * time.Second
		}
	}

	jwt, err := auth.MakeJWT(user.ID, cfg.jwtSecret, expiresInSecondsDuration)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error making JWT", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID: user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email: user.Email,
		},
		Token: jwt,
	})
}