package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/kai-xlr/neo_chirpy/internal/auth"
	"github.com/kai-xlr/neo_chirpy/internal/database"
)

// handlerUsers dispatches user-related requests based on HTTP method
func (a *apiConfig) handlerUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		a.handlerUsersCreate(w, r)
	case http.MethodPut:
		a.handlerUsersUpdate(w, r)
	default:
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
	}
}

// handlerUsersCreate handles user registration requests
func (a *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	// Note: method checking is handled by handlerUsers dispatcher

	// Parse request body
	var params userRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	// Validate input
	if err := validateUserCreationRequest(params); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Hash password for secure storage
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	// Create user in database
	user, err := a.db.CreateUserWithPassword(r.Context(), database.CreateUserWithPasswordParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	// Return user response (excluding sensitive data)
	respondWithJSON(w, http.StatusCreated, userResponse{
		User: User{
			ID:          user.ID,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			Email:       user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
	})
}

// handlerLogin handles user authentication requests
func (a *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	// Parse request body
	var params loginRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	// Validate input
	if err := validateLoginRequest(params); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Authenticate user (validates both email and password)
	user, err := a.authenticateUser(r.Context(), params.Email, params.Password)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, auth.ErrInvalidCredentials.Error(), err)
		return
	}

	// Create access token (JWT) that expires in 1 hour
	accessToken, err := auth.MakeJWT(user.ID, a.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create access token", err)
		return
	}

	// Create refresh token that expires in 60 days
	refreshTokenString, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create refresh token", err)
		return
	}

	// Store refresh token in database
	refreshTokenExpiry := time.Now().UTC().Add(60 * 24 * time.Hour) // 60 days
	_, err = a.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshTokenString,
		UserID:    user.ID,
		ExpiresAt: refreshTokenExpiry,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't store refresh token", err)
		return
	}

	// Return authentication response with both tokens
	respondWithJSON(w, http.StatusOK, loginResponse{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		IsChirpyRed:  user.IsChirpyRed,
		Token:        accessToken,
		RefreshToken: refreshTokenString,
	})
}

// handlerRefresh handles POST /api/refresh requests.
// It validates a refresh token and returns a new access token.
// Returns 200 OK with new access token, 401 for invalid/expired/revoked tokens.
func (a *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	// Extract refresh token from Authorization header
	refreshTokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	// Get user from refresh token (validates token exists, not expired, not revoked)
	user, err := a.db.GetUserFromRefreshToken(r.Context(), refreshTokenString)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid or expired refresh token", err)
		return
	}

	// Create new access token that expires in 1 hour
	accessToken, err := auth.MakeJWT(user.ID, a.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create access token", err)
		return
	}

	// Return new access token
	respondWithJSON(w, http.StatusOK, refreshResponse{
		Token: accessToken,
	})
}

// handlerRevoke handles POST /api/revoke requests.
// It revokes a refresh token by setting revoked_at to current timestamp.
// Returns 204 No Content on success, 401 for invalid tokens.
func (a *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	// Extract refresh token from Authorization header
	refreshTokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	// Revoke the refresh token
	_, err = a.db.RevokeRefreshToken(r.Context(), refreshTokenString)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token", err)
		return
	}

	// Return 204 No Content for successful revocation
	w.WriteHeader(http.StatusNoContent)
}

// handlerUsersUpdate handles PUT /api/users requests.
// It allows users to update their own email and password.
// Requires a valid JWT access token.
// Returns 200 OK with updated user info, 401 for invalid token, 400 for validation errors.
func (a *apiConfig) handlerUsersUpdate(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPut) {
		return
	}

	// Extract and validate JWT token
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	userID, err := auth.ValidateJWT(tokenString, a.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	// Parse request body
	var params userUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	// Validate input
	if err := validateUserUpdateRequest(params); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Hash the new password for secure storage
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	// Update user in database
	updatedUser, err := a.db.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userID,
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update user", err)
		return
	}

	// Return updated user response (excluding sensitive data)
	respondWithJSON(w, http.StatusOK, userResponse{
		User: User{
			ID:          updatedUser.ID,
			CreatedAt:   updatedUser.CreatedAt,
			UpdatedAt:   updatedUser.UpdatedAt,
			Email:       updatedUser.Email,
			IsChirpyRed: updatedUser.IsChirpyRed,
		},
	})
}
