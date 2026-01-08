package auth

import (
	"errors"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
)

// Common authentication errors
var (
	ErrPasswordEmpty      = errors.New("password cannot be empty")
	ErrPasswordMismatch   = errors.New("password does not match")
	ErrPasswordNotSet     = errors.New("user has not set a password")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

// HashPassword creates a secure hash from a plain text password
// Uses Argon2id, which is the recommended password hashing algorithm
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", ErrPasswordEmpty
	}

	// Create hash with secure default parameters
	hashedPassword, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}

	return hashedPassword, nil
}

// VerifyPassword checks if a plain text password matches a stored hash
// Returns an error if the passwords don't match or the hash is invalid
func VerifyPassword(plainPassword, hashedPassword string) error {
	if plainPassword == "" {
		return ErrPasswordEmpty
	}

	if hashedPassword == "" || hashedPassword == "unset" {
		return ErrPasswordNotSet
	}

	// Compare password with hash
	match, err := argon2id.ComparePasswordAndHash(plainPassword, hashedPassword)
	if err != nil {
		return err // Hash format error
	}

	if !match {
		return ErrPasswordMismatch
	}

	return nil
}

// CreateAccessToken generates a JWT token for a user
// TODO: Replace with proper JWT implementation
func CreateAccessToken(userID uuid.UUID) (string, error) {
	// Temporary implementation - just returns user ID as token
	// In production, use a proper JWT library like github.com/golang-jwt/jwt
	return userID.String(), nil
}
