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
	chirpID, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 404, "invalid chirpID")
		return
	}
	dbChirp, err := cfg.db.GetChirp(req.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "fetching chirp server error")
		return
	}
	respondWithJSON(w, http.StatusOK, ChirpResponse{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	})
}

func (cfg *apiConfig) handlerChirpsGetAll(w http.ResponseWriter, req *http.Request) {
	//from GET /api/chirps
	//can be queried
	var err error
	var dbChirps []database.Chirp

	author := false
	authorParam := req.URL.Query().Get("author_id")
	if authorParam != "" {
		author = true
	}

	sort := "ASC"
	sortParam := req.URL.Query().Get("sort")
	if strings.ToUpper(sortParam) == "DESC" {
		sort = "DESC"
	}

	// baseQuery := "SELECT * FROM chirps"
	// if author {
	// 	baseQuery += "WHERE user_id = $1"
	// }
	// baseQuery += "ORDER BY created_at $2"

	if author {
		var userID uuid.UUID
		userID, err = uuid.Parse(authorParam)
		if err != nil {
			respondWithError(w, 401, "parsing author id gone wrong")
			return
		}
		if sort == "ASC" {
			dbChirps, err = cfg.db.GetAllChirpsFromAuthorASC(req.Context(), userID)
		} else {
			dbChirps, err = cfg.db.GetAllChirpsFromAuthorDESC(req.Context(), userID)
		}
	} else {
		if sort == "ASC" {
			dbChirps, err = cfg.db.ListChirpsASC(req.Context())
		} else {
			dbChirps, err = cfg.db.ListChirpsDESC(req.Context())
		}
	}

	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	response := []ChirpResponse{}

	for _, dbChirp := range dbChirps {
		response = append(response, ChirpResponse{
			ID:        dbChirp.ID,
			Body:      dbChirp.Body,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			UserID:    dbChirp.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, response)
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
	dbchirp, err := cfg.db.CreateChirp(req.Context(), database.CreateChirpParams{
		Body:   validatedChirp,
		UserID: userID,
	})
	if err != nil {
		respondWithError(w, 500, "chirp creation db error")
	}
	respondWithJSON(w, http.StatusCreated, ChirpResponse{
		ID:        dbchirp.ID,
		CreatedAt: dbchirp.CreatedAt,
		UpdatedAt: dbchirp.UpdatedAt,
		Body:      dbchirp.Body,
		UserID:    dbchirp.UserID,
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

type ChirpResponse struct {
	ID        uuid.UUID `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `json:"user_id"`
}
