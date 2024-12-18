package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/chichigami/chirpy/internal/auth"
	"github.com/chichigami/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUsersUpdate(w http.ResponseWriter, req *http.Request) {
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

	param := parameter{}
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&param); err != nil {
		respondWithError(w, 400, err.Error())
		return
	}

	newPassword, err := auth.HashPassword(param.Password)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	cfg.db.UpdateEmailandPassword(req.Context(), database.UpdateEmailandPasswordParams{
		ID:             userID,
		Email:          param.Email,
		HashedPassword: newPassword,
	})
	type User struct {
		ID            uuid.UUID `json:"id"`
		CreatedAt     time.Time `json:"created_at"`
		UpdatedAt     time.Time `json:"updated_at"`
		Email         string    `json:"email"`
		Is_Chirpy_Red bool      `json:"is_chirpy_red"`
	}
	dbUser, err := cfg.db.GetUserByEmail(req.Context(), param.Email)
	if err != nil {
		respondWithError(w, 500, "fetching user failed")
		return
	}
	respondWithJSON(w, 200, User{
		ID:            dbUser.ID,
		CreatedAt:     dbUser.CreatedAt,
		UpdatedAt:     dbUser.UpdatedAt,
		Email:         dbUser.Email,
		Is_Chirpy_Red: dbUser.IsChirpyRed.Bool,
	})
}

func (cfg *apiConfig) handlerUsersLogin(w http.ResponseWriter, req *http.Request) {
	type User struct {
		ID            uuid.UUID `json:"id"`
		CreatedAt     time.Time `json:"created_at"`
		UpdatedAt     time.Time `json:"updated_at"`
		Email         string    `json:"email"`
		Token         string    `json:"token"`
		Refresh_Token string    `json:"refresh_token"`
		Is_Chirpy_Red bool      `json:"is_chirpy_red"`
	}
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

	const maxTokenDuration = time.Hour

	userToken, err := auth.MakeJWT(dbUser.ID, cfg.jwtSecret, maxTokenDuration)
	if err != nil {
		respondWithError(w, 500, "access token generation failed")
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, 500, "refresh token generation failed")
		return
	}
	refreshTokenParam := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    dbUser.ID,
		ExpiresAt: sql.NullTime{Time: time.Now().Add(time.Hour * 24 * 60), Valid: true},
		RevokedAt: sql.NullTime{},
	}
	cfg.db.CreateRefreshToken(req.Context(), refreshTokenParam)

	respondWithJSON(w, http.StatusOK, User{
		dbUser.ID,
		dbUser.CreatedAt,
		dbUser.UpdatedAt,
		dbUser.Email,
		userToken,
		refreshToken,
		dbUser.IsChirpyRed.Bool,
	})
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, req *http.Request) {
	//called from POST /api/users
	type User struct {
		ID            uuid.UUID `json:"id"`
		CreatedAt     time.Time `json:"created_at"`
		UpdatedAt     time.Time `json:"updated_at"`
		Email         string    `json:"email"`
		Is_Chirpy_Red bool      `json:"is_chirpy_red"`
	}
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
		IsChirpyRed:    sql.NullBool{Bool: false, Valid: false},
	})

	if dbErr != nil {
		respondWithError(w, http.StatusInternalServerError, dbErr.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, User{
		ID:            dbUser.ID,
		CreatedAt:     dbUser.CreatedAt,
		UpdatedAt:     dbUser.UpdatedAt,
		Email:         dbUser.Email,
		Is_Chirpy_Red: dbUser.IsChirpyRed.Bool,
	})
}

type parameter struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}
