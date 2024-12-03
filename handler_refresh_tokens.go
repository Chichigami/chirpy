package main

import (
	"net/http"
	"time"

	"github.com/chichigami/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRevokeRefreshToken(w http.ResponseWriter, req *http.Request) {
	reqToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, 401, "refresh token expired or does not exist")
		return
	}
	cfg.db.RevokeRefreshToken(req.Context(), reqToken)
	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, req *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	reqToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	dbUser, err := cfg.db.GetUserFromRefreshToken(req.Context(), reqToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "refresh token expired or does not exist")
		return
	}
	if dbUser.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "refresh token revoked")
	}
	jwtToken, err := auth.MakeJWT(dbUser.UserID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, 500, "access token generation failed")
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Token: jwtToken,
	})
}
