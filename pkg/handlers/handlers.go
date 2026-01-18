package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/kai-xlr/neo_chirpy/internal/database"
	"github.com/kai-xlr/neo_chirpy/pkg/types"
)

// RequireMethod validates the HTTP method and returns false if invalid
func RequireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}

type errorResponse struct {
	Error string `json:"error"`
}

// RespondWithError sends an error response in JSON format
func RespondWithError(w http.ResponseWriter, code int, msg string, err error) {
	// Log the actual error for debugging purposes
	if err != nil {
		log.Println(err)
	}
	// Log 5XX errors specifically as they indicate server problems
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	// Send error response in JSON format
	RespondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

// RespondWithJSON sends a JSON response
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", types.ContentTypeJSON)
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

// BuildChirpResponse converts a database chirp to API response format
func BuildChirpResponse(dbChirp database.Chirp) types.ChirpCreateResponse {
	return types.ChirpCreateResponse{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}
}

// BuildChirpListResponse converts a slice of database chirps to API response format
func BuildChirpListResponse(dbChirps []database.Chirp) []types.ChirpCreateResponse {
	response := make([]types.ChirpCreateResponse, len(dbChirps))
	for chirpIdx, chirp := range dbChirps {
		response[chirpIdx] = BuildChirpResponse(chirp)
	}
	return response
}

// PathMatch checks if the path starts with the given prefix
func PathMatch(path, prefix string) bool {
	return len(path) > len(prefix) && path[:len(prefix)] == prefix
}

// ExtractIDFromPath extracts the ID part from a path like "/api/chirps/{id}"
func ExtractIDFromPath(path, prefix string) string {
	if len(path) <= len(prefix) {
		return ""
	}
	return path[len(prefix):]
}
