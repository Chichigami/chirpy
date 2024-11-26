package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func handlerValidateChirp(w http.ResponseWriter, req *http.Request) {
	const chirpMaxLength = 140
	const longChirp = "Chirp is too long"
	type parameter struct {
		Body string `json:"body"`
	}

	type returnVal struct {
		Cleaned_body string `json:"cleaned_body"`
	}

	type errorVal struct {
		Error string `json:"error"`
	}

	decoder := json.NewDecoder(req.Body)
	param := parameter{}
	if decodeErr := decoder.Decode(&param); decodeErr != nil {
		data, _ := json.Marshal(errorVal{Error: decodeErr.Error()})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write(data)
		return
	}

	if len(param.Body) > chirpMaxLength {
		data, _ := json.Marshal(errorVal{Error: longChirp})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write(data)
		return
	}
	profaneCleaner(param.Body)
	respBody := returnVal{
		Cleaned_body: profaneCleaner(param.Body),
	}
	data, encodeErr := json.Marshal(respBody)
	if encodeErr != nil {
		log.Printf("Error marshaling JSON %s", encodeErr)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
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
