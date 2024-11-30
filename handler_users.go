package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, req *http.Request) {
	//called from POST /api/users
	type parameter struct {
		Email string `json:"email"`
	}

	param := parameter{}
	decoder := json.NewDecoder(req.Body)
	if decodeErr := decoder.Decode(&param); decodeErr != nil {
		respondWithError(w, 400, decodeErr.Error())
		return
	}
	newUser, dbErr := cfg.db.CreateUser(req.Context(), param.Email)
	if dbErr != nil {
		respondWithError(w, http.StatusInternalServerError, dbErr.Error())
	}

	respondWithJSON(w, http.StatusCreated, User{
		newUser.ID,
		newUser.CreatedAt,
		newUser.UpdatedAt,
		newUser.Email,
	})
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}
