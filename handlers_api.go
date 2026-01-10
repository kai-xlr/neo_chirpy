// Package handlers_api contains HTTP handlers for the chirpy API.
// It provides endpoints for chirp management, user operations, and health checks.
package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/kai-xlr/neo_chirpy/internal/auth"
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
// It validates the JWT token, chirp body, filters profanity, and stores the chirp in the database.
// Returns 201 Created on success, 401 for invalid token, 400 for validation errors, 500 for server errors.
func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	// Extract and validate JWT token
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	userID, err := auth.ValidateJWT(tokenString, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
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
		UserID: userID,
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

// handlerChirpByID handles GET /api/chirps/{id} requests.
// It retrieves a specific chirp by its ID from the database.
// Returns 200 OK with the chirp on success, 404 if not found, 500 for server errors.
func (cfg *apiConfig) handlerChirpByID(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	// Extract chirp ID from URL path
	path := r.URL.Path
	if !pathMatch(path, "/api/chirps/") {
		respondWithError(w, http.StatusNotFound, "404 page not found", nil)
		return
	}

	// Extract ID from path "/api/chirps/{id}"
	chirpID := extractIDFromPath(path, "/api/chirps/")
	if chirpID == "" {
		respondWithError(w, http.StatusBadRequest, "Chirp ID is required", nil)
		return
	}

	// Parse UUID
	parsedID, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID format", err)
		return
	}

	// Retrieve chirp from database
	dbChirp, err := cfg.db.GetChirpByID(r.Context(), parsedID)
	if err != nil {
		if err.Error() == "no rows in result set" || err.Error() == "sql: no rows in result set" {
			respondWithError(w, http.StatusNotFound, "404 page not found", nil)
		} else {
			respondWithError(w, http.StatusInternalServerError, ErrMsgRetrieveChirp, err)
		}
		return
	}

	respondWithJSON(w, http.StatusOK, buildChirpResponse(dbChirp))
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

// pathMatch checks if the path starts with the given prefix
func pathMatch(path, prefix string) bool {
	return len(path) > len(prefix) && path[:len(prefix)] == prefix
}

// extractIDFromPath extracts the ID part from a path like "/api/chirps/{id}"
func extractIDFromPath(path, prefix string) string {
	if len(path) <= len(prefix) {
		return ""
	}
	return path[len(prefix):]
}
