package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/chichigami/chirpy/internal/auth"
	"github.com/chichigami/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerChirpsDeleteID(w http.ResponseWriter, req *http.Request) {
	chirpID, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 404, "invalid chirpID")
		return
	}

	jwtToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	userID, err := auth.ValidateJWT(jwtToken, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	chirp, err := cfg.db.GetChirp(req.Context(), chirpID)
	if err != nil {
		respondWithError(w, 404, "chirp is not found")
		return
	}
	if chirp.UserID != userID {
		respondWithError(w, 403, "authorization not valid")
		return
	}
	cfg.db.DeleteChirp(req.Context(), chirpID)
	w.WriteHeader(204)
}

func (cfg *apiConfig) handlerChirpsGetID(w http.ResponseWriter, req *http.Request) {
	chripID, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 404, "invalid chirpID")
		return
	}
	dbChirp, err := cfg.db.GetChirp(req.Context(), chripID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "fetching chirp server error")
		return
	}
	respondWithJSON(w, http.StatusOK, chirpResource{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		User_ID:   dbChirp.UserID,
	})
}

func (cfg *apiConfig) handlerChirpsGetAll(w http.ResponseWriter, req *http.Request) {
	//from GET /api/chirps
	//can be queried
	query := false
	var err error
	target := req.URL.Query().Get("author_id")
	if target != "" {
		query = true
	}

	var dbChirps []database.Chirp

	if query {
		var userID uuid.UUID
		userID, err = uuid.Parse(target)
		if err != nil {
			respondWithError(w, 401, "parsing author id gone wrong")
			return
		}
		dbChirps, err = cfg.db.GetAllChirpsFromAuthor(req.Context(), userID)
	} else {
		dbChirps, err = cfg.db.ListChrips(req.Context())
	}

	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, dbChirps)
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, req *http.Request) {
	//POST /api/chirps
	userToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}
	userID, err := auth.ValidateJWT(userToken, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	type parameter struct {
		Body string `json:"body"`
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
		UserID: userID,
	})
	if err != nil {
		respondWithError(w, 500, "chrip creation db error")
	}
	respondWithJSON(w, http.StatusCreated, chirpResource{
		ID:        dbChrip.ID,
		CreatedAt: dbChrip.CreatedAt,
		UpdatedAt: dbChrip.UpdatedAt,
		Body:      dbChrip.Body,
		User_ID:   dbChrip.UserID,
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

type chirpResource struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	User_ID   uuid.UUID `json:"user_id"`
}
