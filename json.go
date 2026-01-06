package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/kai-xlr/neo_chirpy/internal/database"
)

type errorResponse struct {
	Error string `json:"error"`
}

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	// Log the actual error for debugging purposes
	if err != nil {
		log.Println(err)
	}
	// Log 5XX errors specifically as they indicate server problems
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	// Send error response in JSON format
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", ContentTypeJSON)
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

// buildChirpResponse converts a database chirp to API response format
func buildChirpResponse(dbChirp database.Chirp) chirpCreateResponse {
	return chirpCreateResponse{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}
}

// buildChirpListResponse converts a slice of database chirps to API response format
func buildChirpListResponse(dbChirps []database.Chirp) []chirpCreateResponse {
	response := make([]chirpCreateResponse, len(dbChirps))
	for chirpIdx, chirp := range dbChirps {
		response[chirpIdx] = buildChirpResponse(chirp)
	}
	return response
}
