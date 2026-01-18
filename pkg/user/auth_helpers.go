package user

import (
	"context"
	"strings"
	"time"

	"github.com/kai-xlr/neo_chirpy/internal/auth"
	"github.com/kai-xlr/neo_chirpy/internal/database"
	"github.com/kai-xlr/neo_chirpy/pkg/types"
	"github.com/kai-xlr/neo_chirpy/pkg/validation"
)

// Config holds configuration needed for user handlers
type Config struct {
	DB        *database.Queries
	JWTSecret string
}

// validateLoginRequest checks if login request is valid
func validateLoginRequest(req types.LoginRequest) error {
	if req.Email == "" {
		return validation.ErrEmailEmpty
	}
	if req.Password == "" {
		return auth.ErrPasswordEmpty
	}
	return nil
}

// validateUserCreationRequest checks if user creation request is valid
func validateUserCreationRequest(req types.UserRequest) error {
	if err := validation.ValidateEmail(req.Email); err != nil {
		return err
	}
	if req.Password == "" {
		return auth.ErrPasswordEmpty
	}
	return nil
}

// validateUserUpdateRequest checks if user update request is valid
func validateUserUpdateRequest(req types.UserUpdateRequest) error {
	if err := validation.ValidateEmail(req.Email); err != nil {
		return err
	}
	if strings.TrimSpace(req.Password) == "" {
		return auth.ErrPasswordEmpty
	}
	return nil
}

// authenticateUser verifies user credentials and returns user if valid
func (cfg *Config) authenticateUser(ctx context.Context, email, password string) (database.User, error) {
	// Get user from database
	user, err := cfg.DB.GetUserByEmail(ctx, email)
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

// createTokens creates both access and refresh tokens for a user
func (cfg *Config) createTokens(ctx context.Context, user database.User) (string, string, error) {
	// Create access token (JWT) that expires in 1 hour
	accessToken, err := auth.MakeJWT(user.ID, cfg.JWTSecret, time.Hour)
	if err != nil {
		return "", "", err
	}

	// Create refresh token that expires in 60 days
	refreshTokenString, err := auth.MakeRefreshToken()
	if err != nil {
		return "", "", err
	}

	// Store refresh token in database
	refreshTokenExpiry := time.Now().UTC().Add(60 * 24 * time.Hour) // 60 days
	_, err = cfg.DB.CreateRefreshToken(ctx, database.CreateRefreshTokenParams{
		Token:     refreshTokenString,
		UserID:    user.ID,
		ExpiresAt: refreshTokenExpiry,
	})
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshTokenString, nil
}
