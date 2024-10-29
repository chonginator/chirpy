package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/chonginator/chirpy/internal/auth"
	"github.com/chonginator/chirpy/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID uuid.UUID				`json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email string				`json:"email"`
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string		`json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error decoding parameters", err)
		return
	}

	if params.Email == "" {
		err := fmt.Errorf("email field is empty")
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	if params.Password == "" {
		err := fmt.Errorf("password field is empty")
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error hashing password", err)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email: params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating user", err)
		return
	}
	
	respondWithJSON(w, http.StatusCreated, User{
			ID: user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email: user.Email,
		},
	)
}
