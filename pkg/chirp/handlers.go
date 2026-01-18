package chirp

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/google/uuid"
	"github.com/kai-xlr/neo_chirpy/internal/auth"
	"github.com/kai-xlr/neo_chirpy/internal/database"
	"github.com/kai-xlr/neo_chirpy/pkg/handlers"
	"github.com/kai-xlr/neo_chirpy/pkg/types"
	"github.com/kai-xlr/neo_chirpy/pkg/validation"
)

// Config holds the configuration needed for chirp handlers
type Config struct {
	DB        *database.Queries
	JWTSecret string
}

// HandlerCreate handles POST /api/chirps requests.
func (cfg *Config) HandlerCreate(w http.ResponseWriter, r *http.Request) {
	if !handlers.RequireMethod(w, r, http.MethodPost) {
		return
	}

	// Extract and validate JWT token
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		handlers.RespondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	userID, err := auth.ValidateJWT(tokenString, cfg.JWTSecret)
	if err != nil {
		handlers.RespondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	var request types.ChirpCreateRequest
	// Parse JSON from request body into our struct
	decodeErr := json.NewDecoder(r.Body).Decode(&request)
	if decodeErr != nil {
		handlers.RespondWithError(w, http.StatusInternalServerError, types.ErrMsgDecodeParams, decodeErr)
		return
	}

	// Validate chirp body against business rules (max length, empty check)
	if validationErr := validation.ValidateChirpBody(request.Body); validationErr != nil {
		handlers.RespondWithError(w, http.StatusBadRequest, validationErr.Error(), validationErr)
		return
	}

	// Remove profanity from the chirp body
	cleanedBody := CleanChirp(request.Body)

	// Insert chirp into database using generated sqlc code
	createdChirp, dbErr := cfg.DB.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: userID,
	})
	if dbErr != nil {
		handlers.RespondWithError(w, http.StatusInternalServerError, types.ErrMsgCreateChirp, dbErr)
		return
	}

	handlers.RespondWithJSON(w, http.StatusCreated, handlers.BuildChirpResponse(createdChirp))
}

// HandlerGet handles GET /api/chirps requests.
func (cfg *Config) HandlerGet(w http.ResponseWriter, r *http.Request) {
	if !handlers.RequireMethod(w, r, http.MethodGet) {
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
		handlers.RespondWithError(w, http.StatusBadRequest, "Invalid sort parameter. Must be 'asc' or 'desc'", nil)
		return
	}

	var dbChirps []database.Chirp
	var dbErr error

	if authorIDStr != "" {
		// Parse author_id as UUID
		authorID, parseErr := uuid.Parse(authorIDStr)
		if parseErr != nil {
			handlers.RespondWithError(w, http.StatusBadRequest, "Invalid author_id format", parseErr)
			return
		}

		// Retrieve chirps for specific author (ascending order is fine, we'll sort in-memory)
		dbChirps, dbErr = cfg.DB.GetChirpsByAuthorAsc(r.Context(), authorID)
	} else {
		// Retrieve all chirps (ascending order is fine, we'll sort in-memory)
		dbChirps, dbErr = cfg.DB.GetChirpsAsc(r.Context())
	}

	if dbErr != nil {
		handlers.RespondWithError(w, http.StatusInternalServerError, types.ErrMsgRetrieveChirps, dbErr)
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
	response := handlers.BuildChirpListResponse(dbChirps)
	handlers.RespondWithJSON(w, http.StatusOK, response)
}

// HandlerByID handles GET and DELETE /api/chirps/{id} requests.
func (cfg *Config) HandlerByID(w http.ResponseWriter, r *http.Request) {
	// Extract chirp ID from URL path (common to both GET and DELETE)
	path := r.URL.Path
	if !handlers.PathMatch(path, "/api/chirps/") {
		handlers.RespondWithError(w, http.StatusNotFound, "404 page not found", nil)
		return
	}

	// Extract ID from path "/api/chirps/{id}"
	chirpID := handlers.ExtractIDFromPath(path, "/api/chirps/")
	if chirpID == "" {
		handlers.RespondWithError(w, http.StatusBadRequest, "Chirp ID is required", nil)
		return
	}

	// Parse UUID
	parsedID, err := uuid.Parse(chirpID)
	if err != nil {
		handlers.RespondWithError(w, http.StatusBadRequest, "Invalid chirp ID format", err)
		return
	}

	switch r.Method {
	case http.MethodGet:
		cfg.handlerByIDGet(w, r, parsedID)
	case http.MethodDelete:
		cfg.handlerByIDDelete(w, r, parsedID)
	default:
		handlers.RespondWithError(w, http.StatusMethodNotAllowed, types.ErrMsgMethodNotAllowed, nil)
	}
}

// handlerByIDGet handles GET /api/chirps/{id} requests.
func (cfg *Config) handlerByIDGet(w http.ResponseWriter, r *http.Request, chirpID uuid.UUID) {
	// Retrieve chirp from database
	dbChirp, err := cfg.DB.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		if err.Error() == "no rows in result set" || err.Error() == "sql: no rows in result set" {
			handlers.RespondWithError(w, http.StatusNotFound, "404 page not found", nil)
		} else {
			handlers.RespondWithError(w, http.StatusInternalServerError, types.ErrMsgRetrieveChirp, err)
		}
		return
	}

	handlers.RespondWithJSON(w, http.StatusOK, handlers.BuildChirpResponse(dbChirp))
}

// handlerByIDDelete handles DELETE /api/chirps/{id} requests.
func (cfg *Config) handlerByIDDelete(w http.ResponseWriter, r *http.Request, chirpID uuid.UUID) {
	// Extract and validate JWT token
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		handlers.RespondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	userID, err := auth.ValidateJWT(tokenString, cfg.JWTSecret)
	if err != nil {
		handlers.RespondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	// Retrieve chirp from database to verify ownership
	dbChirp, err := cfg.DB.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		if err.Error() == "no rows in result set" || err.Error() == "sql: no rows in result set" {
			handlers.RespondWithError(w, http.StatusNotFound, "404 page not found", nil)
		} else {
			handlers.RespondWithError(w, http.StatusInternalServerError, types.ErrMsgRetrieveChirp, err)
		}
		return
	}

	// Check if user is the author of the chirp
	if dbChirp.UserID != userID {
		handlers.RespondWithError(w, http.StatusForbidden, "Forbidden", nil)
		return
	}

	// Delete chirp from database
	err = cfg.DB.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		handlers.RespondWithError(w, http.StatusInternalServerError, "Couldn't delete chirp", err)
		return
	}

	// Return 204 No Content for successful deletion
	w.WriteHeader(http.StatusNoContent)
}
