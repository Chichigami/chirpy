package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/chichigami/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, req *http.Request) {
	type parameter struct {
		Body    string    `json:"body"`
		User_ID uuid.UUID `json:"user_id"`
	}
	type chirpResource struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		User_ID   uuid.UUID `json:"user_id"`
	}

	param := parameter{}
	decoder := json.NewDecoder(req.Body)
	if decodeErr := decoder.Decode(&param); decodeErr != nil {
		respondWithError(w, 400, decodeErr.Error())
		return
	}
	validatedChirp, err := chirpsValidate(param.Body)
	if err != nil {
		respondWithError(w, 400, err.Error())
	}
	dbChrip, err := cfg.db.CreateChrip(req.Context(), database.CreateChripParams{
		Body:   validatedChirp,
		UserID: param.User_ID,
	})
	if err != nil {
		respondWithError(w, 500, "chrip creation db error")
	}
	respondWithJSON(w, http.StatusCreated, chirpResource{
		dbChrip.ID,
		dbChrip.CreatedAt,
		dbChrip.UpdatedAt,
		dbChrip.Body,
		dbChrip.UserID,
	})
}

func chirpsValidate(chirp string) (string, error) {
	const chirpMaxLength = 140

	if len(chirp) > chirpMaxLength {
		return "", fmt.Errorf("chirp is too long")
	}

	return profaneCleaner(chirp), nil
}

func profaneCleaner(body string) string {
	const replacement = "****"
	profanes := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}

	words := strings.Split(body, " ")
	for i, word := range words {
		if profanes[strings.ToLower(word)] {
			words[i] = replacement
		} else {
			words[i] = word
		}
	}

	return strings.Join(words, " ")
}
