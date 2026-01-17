// Package handlers_api contains HTTP handlers for the chirpy API.
// It provides endpoints for chirp management, user operations, and health checks.
package main

import (
	"encoding/json"
	"net/http"
	"sort"

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
// It retrieves chirps from the database ordered by creation date.
// Accepts optional author_id query parameter to filter by author.
// Accepts optional sort query parameter with values "asc" or "desc" (default: "asc").
// Returns 200 OK with an array of chirps on success, 500 for server errors.
func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	// Check for optional query parameters
	authorIDStr := r.URL.Query().Get("author_id")
	sortParam := r.URL.Query().Get("sort")

	// Default to "asc" if no sort parameter provided
	if sortParam == "" {
		sortParam = "asc"
	}

	// Validate sort parameter
	if sortParam != "asc" && sortParam != "desc" {
		respondWithError(w, http.StatusBadRequest, "Invalid sort parameter. Must be 'asc' or 'desc'", nil)
		return
	}

	var dbChirps []database.Chirp
	var dbErr error

	if authorIDStr != "" {
		// Parse author_id as UUID
		authorID, parseErr := uuid.Parse(authorIDStr)
		if parseErr != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid author_id format", parseErr)
			return
		}

		// Retrieve chirps for specific author (ascending order is fine, we'll sort in-memory)
		dbChirps, dbErr = cfg.db.GetChirpsByAuthorAsc(r.Context(), authorID)
	} else {
		// Retrieve all chirps (ascending order is fine, we'll sort in-memory)
		dbChirps, dbErr = cfg.db.GetChirpsAsc(r.Context())
	}

	if dbErr != nil {
		respondWithError(w, http.StatusInternalServerError, ErrMsgRetrieveChirps, dbErr)
		return
	}

	// Sort chirps in-memory based on the sort parameter
	if sortParam == "desc" {
		sort.Slice(dbChirps, func(i, j int) bool {
			return dbChirps[i].CreatedAt.After(dbChirps[j].CreatedAt)
		})
	} else {
		// sortParam == "asc" - already sorted this way, but ensure consistency
		sort.Slice(dbChirps, func(i, j int) bool {
			return dbChirps[i].CreatedAt.Before(dbChirps[j].CreatedAt)
		})
	}

	// Convert database chirps to API response format using helper function
	response := buildChirpListResponse(dbChirps)
	respondWithJSON(w, http.StatusOK, response)
}

// handlerChirpByID handles GET and DELETE /api/chirps/{id} requests.
// GET: retrieves a specific chirp by its ID from the database.
// DELETE: deletes a chirp by its ID (requires authentication and ownership).
// Returns appropriate status codes for each method and error condition.
func (cfg *apiConfig) handlerChirpByID(w http.ResponseWriter, r *http.Request) {
	// Extract chirp ID from URL path (common to both GET and DELETE)
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

	switch r.Method {
	case http.MethodGet:
		cfg.handlerChirpByIDGet(w, r, parsedID)
	case http.MethodDelete:
		cfg.handlerChirpByIDDelete(w, r, parsedID)
	default:
		respondWithError(w, http.StatusMethodNotAllowed, ErrMsgMethodNotAllowed, nil)
	}
}

// handlerChirpByIDGet handles GET /api/chirps/{id} requests.
// It retrieves a specific chirp by its ID from the database.
// Returns 200 OK with the chirp on success, 404 if not found, 500 for server errors.
func (cfg *apiConfig) handlerChirpByIDGet(w http.ResponseWriter, r *http.Request, chirpID uuid.UUID) {
	// Retrieve chirp from database
	dbChirp, err := cfg.db.GetChirpByID(r.Context(), chirpID)
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

// handlerChirpByIDDelete handles DELETE /api/chirps/{id} requests.
// It validates the JWT token, checks if user is the author of the chirp, and deletes it.
// Returns 204 No Content on success, 401 for invalid token, 403 if not author, 404 if not found, 500 for server errors.
func (cfg *apiConfig) handlerChirpByIDDelete(w http.ResponseWriter, r *http.Request, chirpID uuid.UUID) {
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

	// Retrieve chirp from database to verify ownership
	dbChirp, err := cfg.db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		if err.Error() == "no rows in result set" || err.Error() == "sql: no rows in result set" {
			respondWithError(w, http.StatusNotFound, "404 page not found", nil)
		} else {
			respondWithError(w, http.StatusInternalServerError, ErrMsgRetrieveChirp, err)
		}
		return
	}

	// Check if user is the author of the chirp
	if dbChirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "Forbidden", nil)
		return
	}

	// Delete chirp from database
	err = cfg.db.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't delete chirp", err)
		return
	}

	// Return 204 No Content for successful deletion
	w.WriteHeader(http.StatusNoContent)
}

// handlerChirps routes chirp-related requests to the appropriate handler.
// It supports GET (retrieve all chirps), POST (create new chirp), and DELETE (delete chirp) methods.
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

// handlerPolkaWebhooks handles POST /api/polka/webhooks requests.
// It processes user upgrade events from Polka payment system.
// Returns 204 No Content for successful processing or ignored events, 404 if user not found, 401 if unauthorized.
func (cfg *apiConfig) handlerPolkaWebhooks(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	// Validate API key
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, auth.ErrUnauthorized.Error(), err)
		return
	}

	if apiKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, auth.ErrUnauthorized.Error(), auth.ErrUnauthorized)
		return
	}

	// Parse JSON from request body
	var request webhookRequest
	decodeErr := json.NewDecoder(r.Body).Decode(&request)
	if decodeErr != nil {
		respondWithError(w, http.StatusInternalServerError, ErrMsgDecodeParams, decodeErr)
		return
	}

	// If event is not user.upgraded, respond with 204 (we don't care about other events)
	if request.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Upgrade user to Chirpy Red
	_, err = cfg.db.UpgradeUserToChirpyRed(r.Context(), request.Data.UserID)
	if err != nil {
		if err.Error() == "no rows in result set" || err.Error() == "sql: no rows in result set" {
			respondWithError(w, http.StatusNotFound, "User not found", err)
		} else {
			respondWithError(w, http.StatusInternalServerError, "Couldn't upgrade user", err)
		}
		return
	}

	// Return 204 No Content for successful upgrade
	w.WriteHeader(http.StatusNoContent)
}
