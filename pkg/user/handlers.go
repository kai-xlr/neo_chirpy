package user

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/kai-xlr/neo_chirpy/internal/auth"
	"github.com/kai-xlr/neo_chirpy/internal/database"
	"github.com/kai-xlr/neo_chirpy/pkg/handlers"
	"github.com/kai-xlr/neo_chirpy/pkg/types"
)

// HandlerUsers dispatches user-related requests based on HTTP method
func (cfg *Config) HandlerUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		cfg.handlerUsersCreate(w, r)
	case http.MethodPut:
		cfg.handlerUsersUpdate(w, r)
	default:
		handlers.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
	}
}

// handlerUsersCreate handles user registration requests
func (cfg *Config) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var params types.UserRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		handlers.RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	// Validate input
	if err := validateUserCreationRequest(params); err != nil {
		handlers.RespondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Hash password for secure storage
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		handlers.RespondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	// Create user in database
	user, err := cfg.DB.CreateUserWithPassword(r.Context(), database.CreateUserWithPasswordParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		handlers.RespondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	// Return user response (excluding sensitive data)
	handlers.RespondWithJSON(w, http.StatusCreated, types.UserResponse{
		User: types.User{
			ID:          user.ID,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			Email:       user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
	})
}

// HandlerLogin handles user authentication requests
func (cfg *Config) HandlerLogin(w http.ResponseWriter, r *http.Request) {
	if !handlers.RequireMethod(w, r, http.MethodPost) {
		return
	}

	// Parse request body
	var params types.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		handlers.RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	// Validate input
	if err := validateLoginRequest(params); err != nil {
		handlers.RespondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Authenticate user (validates both email and password)
	user, err := cfg.authenticateUser(r.Context(), params.Email, params.Password)
	if err != nil {
		handlers.RespondWithError(w, http.StatusUnauthorized, auth.ErrInvalidCredentials.Error(), err)
		return
	}

	// Create tokens
	accessToken, refreshTokenString, err := cfg.createTokens(r.Context(), user)
	if err != nil {
		handlers.RespondWithError(w, http.StatusInternalServerError, "Couldn't create tokens", err)
		return
	}

	// Return authentication response with both tokens
	handlers.RespondWithJSON(w, http.StatusOK, types.LoginResponse{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		IsChirpyRed:  user.IsChirpyRed,
		Token:        accessToken,
		RefreshToken: refreshTokenString,
	})
}

// HandlerRefresh handles POST /api/refresh requests
func (cfg *Config) HandlerRefresh(w http.ResponseWriter, r *http.Request) {
	if !handlers.RequireMethod(w, r, http.MethodPost) {
		return
	}

	// Extract refresh token from Authorization header
	refreshTokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		handlers.RespondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	// Get user from refresh token (validates token exists, not expired, not revoked)
	user, err := cfg.DB.GetUserFromRefreshToken(r.Context(), refreshTokenString)
	if err != nil {
		handlers.RespondWithError(w, http.StatusUnauthorized, "Invalid or expired refresh token", err)
		return
	}

	// Create new access token that expires in 1 hour
	accessToken, err := auth.MakeJWT(user.ID, cfg.JWTSecret, time.Hour)
	if err != nil {
		handlers.RespondWithError(w, http.StatusInternalServerError, "Couldn't create access token", err)
		return
	}

	// Return new access token
	handlers.RespondWithJSON(w, http.StatusOK, types.RefreshResponse{
		Token: accessToken,
	})
}

// HandlerRevoke handles POST /api/revoke requests
func (cfg *Config) HandlerRevoke(w http.ResponseWriter, r *http.Request) {
	if !handlers.RequireMethod(w, r, http.MethodPost) {
		return
	}

	// Extract refresh token from Authorization header
	refreshTokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		handlers.RespondWithError(w, http.StatusUnauthorized, "Invalid token", err)
		return
	}

	// Revoke the refresh token
	_, err = cfg.DB.RevokeRefreshToken(r.Context(), refreshTokenString)
	if err != nil {
		handlers.RespondWithError(w, http.StatusUnauthorized, "Invalid refresh token", err)
		return
	}

	// Return 204 No Content for successful revocation
	w.WriteHeader(http.StatusNoContent)
}

// handlerUsersUpdate handles PUT /api/users requests
func (cfg *Config) handlerUsersUpdate(w http.ResponseWriter, r *http.Request) {
	if !handlers.RequireMethod(w, r, http.MethodPut) {
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

	// Parse request body
	var params types.UserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		handlers.RespondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	// Validate input
	if err := validateUserUpdateRequest(params); err != nil {
		handlers.RespondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Hash the new password for secure storage
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		handlers.RespondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	// Update user in database
	updatedUser, err := cfg.DB.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userID,
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		handlers.RespondWithError(w, http.StatusInternalServerError, "Couldn't update user", err)
		return
	}

	// Return updated user response (excluding sensitive data)
	handlers.RespondWithJSON(w, http.StatusOK, types.UserResponse{
		User: types.User{
			ID:          updatedUser.ID,
			CreatedAt:   updatedUser.CreatedAt,
			UpdatedAt:   updatedUser.UpdatedAt,
			Email:       updatedUser.Email,
			IsChirpyRed: updatedUser.IsChirpyRed,
		},
	})
}
