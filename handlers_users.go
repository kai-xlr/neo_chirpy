package main

import (
	"encoding/json"
	"net/http"

	"github.com/kai-xlr/neo_chirpy/internal/auth"
	"github.com/kai-xlr/neo_chirpy/internal/database"
)

// handlerUsersCreate handles user registration requests
func (a *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

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
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
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

	// Create access token
	token, err := auth.CreateAccessToken(user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create token", err)
		return
	}

	// Return authentication response
	respondWithJSON(w, http.StatusOK, loginResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     token,
	})
}
