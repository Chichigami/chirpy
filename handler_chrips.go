package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func handlerChripsValidate(w http.ResponseWriter, req *http.Request) {
	const chirpMaxLength = 140
	const longChirp = "Chirp is too long"
	type parameter struct {
		Body string `json:"body"`
	}

	type returnVal struct {
		Cleaned_body string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(req.Body)
	param := parameter{}
	if decodeErr := decoder.Decode(&param); decodeErr != nil {
		respondWithError(w, 400, decodeErr.Error())
		return
	}

	if len(param.Body) > chirpMaxLength {
		respondWithError(w, 400, longChirp)
		return
	}

	respBody := returnVal{
		Cleaned_body: profaneCleaner(param.Body),
	}
	respondWithJSON(w, http.StatusOK, respBody)
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
