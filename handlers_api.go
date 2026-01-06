// Package handlers_api contains HTTP handlers for the chirpy API.
// It provides endpoints for chirp management, user operations, and health checks.
package main

import (
	"encoding/json"
	"net/http"

	"github.com/kai-xlr/neo_chirpy/internal/database"
)

// handlerReadiness responds to GET /api/healthz with a simple "OK" message.
// It's used by load balancers and monitoring systems to verify the service is running.
func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	w.Header().Set("Content-Type", ContentTypeTextPlain)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

// handlerChirpsCreate handles POST /api/chirps requests.
// It validates the chirp body, filters profanity, and stores the chirp in the database.
// Returns 201 Created on success, 400 for validation errors, 500 for server errors.
func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var request chirpCreateRequest
	// Parse JSON from request body into our struct
	decodeErr := json.NewDecoder(r.Body).Decode(&request)
	if decodeErr != nil {
		respondWithError(w, http.StatusInternalServerError, ErrMsgDecodeParams, decodeErr)
		return
	}

	// Validate chirp body against business rules (max length, empty check)
	if validationErr := ValidateChirpBody(request.Body); validationErr != nil {
		respondWithError(w, http.StatusBadRequest, validationErr.Error(), validationErr)
		return
	}

	// Remove profanity from the chirp body using the sanitize module
	cleanedBody := cleanChirp(request.Body)

	// Insert chirp into database using generated sqlc code
	createdChirp, dbErr := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: request.UserID,
	})
	if dbErr != nil {
		respondWithError(w, http.StatusInternalServerError, ErrMsgCreateChirp, dbErr)
		return
	}

	respondWithJSON(w, http.StatusCreated, buildChirpResponse(createdChirp))
}

// handlerChirpsGet handles GET /api/chirps requests.
// It retrieves all chirps from the database ordered by creation date (oldest first).
// Returns 200 OK with an array of chirps on success, 500 for server errors.
func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	// Retrieve all chirps from database, ordered by creation date (oldest first)
	dbChirps, dbErr := cfg.db.GetChirpsAsc(r.Context())
	if dbErr != nil {
		respondWithError(w, http.StatusInternalServerError, ErrMsgRetrieveChirps, dbErr)
		return
	}

	// Convert database chirps to API response format using helper function
	response := buildChirpListResponse(dbChirps)
	respondWithJSON(w, http.StatusOK, response)
}

// handlerChirps routes chirp-related requests to the appropriate handler.
// It supports GET (retrieve all chirps) and POST (create new chirp) methods.
// Returns 405 Method Not Allowed for unsupported HTTP methods.
func (cfg *apiConfig) handlerChirps(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg.handlerChirpsGet(w, r)
	case http.MethodPost:
		cfg.handlerChirpsCreate(w, r)
	default:
		respondWithError(w, http.StatusMethodNotAllowed, ErrMsgMethodNotAllowed, nil)
	}
}
