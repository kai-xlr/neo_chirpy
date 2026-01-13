package main

import (
	"context"
	"strings"

	"github.com/kai-xlr/neo_chirpy/internal/auth"
	"github.com/kai-xlr/neo_chirpy/internal/database"
)

// validateLoginRequest checks if the login request is valid
func validateLoginRequest(req loginRequest) error {
	if req.Email == "" {
		return ErrEmailEmpty
	}
	if req.Password == "" {
		return auth.ErrPasswordEmpty
	}
	return nil
}

// validateUserCreationRequest checks if user creation request is valid
func validateUserCreationRequest(req userRequest) error {
	if err := ValidateEmail(req.Email); err != nil {
		return err
	}
	if req.Password == "" {
		return auth.ErrPasswordEmpty
	}
	return nil
}

// validateUserUpdateRequest checks if user update request is valid
func validateUserUpdateRequest(req userUpdateRequest) error {
	if err := ValidateEmail(req.Email); err != nil {
		return err
	}
	if strings.TrimSpace(req.Password) == "" {
		return auth.ErrPasswordEmpty
	}
	return nil
}

// authenticateUser verifies user credentials and returns user if valid
func (a *apiConfig) authenticateUser(ctx context.Context, email, password string) (database.User, error) {
	// Get user from database
	user, err := a.db.GetUserByEmail(ctx, email)
	if err != nil {
		return database.User{}, auth.ErrInvalidCredentials
	}

	// Verify password
	err = auth.VerifyPassword(password, user.HashedPassword)
	if err != nil {
		return database.User{}, auth.ErrInvalidCredentials
	}

	return user, nil
}
