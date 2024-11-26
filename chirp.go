package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func handlerValidateChirp(w http.ResponseWriter, req *http.Request) {
	const chirpMaxLength = 140
	const longChirp = "Chirp is too long"
	type parameter struct {
		Body string `json:"body"`
	}

	type returnVal struct {
		Valid bool `json:"valid"`
	}

	type errorVal struct {
		Error string `json:"error"`
	}

	decoder := json.NewDecoder(req.Body)
	param := parameter{}
	decodeErr := decoder.Decode(&param)
	if decodeErr != nil {
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

	respBody := returnVal{}
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
