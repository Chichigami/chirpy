package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/chichigami/chirpy/internal/auth"
	"github.com/chichigami/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUsersLogin(w http.ResponseWriter, req *http.Request) {
	param := parameter{}
	decoder := json.NewDecoder(req.Body)
	if decodeErr := decoder.Decode(&param); decodeErr != nil {
		respondWithError(w, 500, decodeErr.Error())
		return
	}
	dbUser, err := cfg.db.GetUserByEmail(req.Context(), param.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}
	if err := auth.CheckPasswordHash(dbUser.HashedPassword, param.Password); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		dbUser.ID,
		dbUser.CreatedAt,
		dbUser.UpdatedAt,
		dbUser.Email,
	})
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, req *http.Request) {
	//called from POST /api/users
	param := parameter{}
	decoder := json.NewDecoder(req.Body)
	if decodeErr := decoder.Decode(&param); decodeErr != nil {
		respondWithError(w, 400, decodeErr.Error())
		return
	}
	hashedPass, err := auth.HashPassword(param.Password)
	if err != nil {
		respondWithError(w, 500, "failed to hash password")
		return
	}

	dbUser, dbErr := cfg.db.CreateUser(req.Context(), database.CreateUserParams{
		Email:          param.Email,
		HashedPassword: hashedPass,
	})

	if dbErr != nil {
		respondWithError(w, http.StatusInternalServerError, dbErr.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, User{
		dbUser.ID,
		dbUser.CreatedAt,
		dbUser.UpdatedAt,
		dbUser.Email,
	})
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type parameter struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}
