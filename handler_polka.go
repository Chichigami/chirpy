package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/chichigami/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerPolkaWebhook(w http.ResponseWriter, req *http.Request) {
	type parameter struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	param := parameter{}
	decoder := json.NewDecoder(req.Body)
	if decodeErr := decoder.Decode(&param); decodeErr != nil {
		respondWithError(w, 400, decodeErr.Error())
		return
	}
	if param.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}

	userID, err := uuid.Parse(param.Data.UserID)
	if err != nil {
		respondWithError(w, 500, "problem with parsing user id")
		return
	}
	err = cfg.db.UpdateMembership(req.Context(), database.UpdateMembershipParams{
		ID:          userID,
		IsChirpyRed: sql.NullBool{Bool: true, Valid: true},
	})
	if err != nil {
		respondWithError(w, 404, "user cannot be found")
		return
	}
	w.WriteHeader(204)
	w.Write([]byte{})
}
