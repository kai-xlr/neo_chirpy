package auth

import (
	"errors"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Common authentication errors
var (
	ErrPasswordEmpty      = errors.New("password cannot be empty")
	ErrPasswordMismatch   = errors.New("password does not match")
	ErrPasswordNotSet     = errors.New("user has not set a password")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid token")
	ErrExpiredToken       = errors.New("token has expired")
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

// MakeJWT creates a JWT token for a user with the specified secret and expiration time
func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	now := time.Now().UTC()

	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// ValidateJWT checks if a JWT token is valid and returns the user ID
func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(tokenSecret), nil
	})

	if err != nil {
		// Check for specific JWT library errors
		errStr := err.Error()
		if errStr == "token is expired" ||
			errStr == "token has invalid claims: token is expired" {
			return uuid.Nil, ErrExpiredToken
		}
		// Handle various signature invalid error formats
		if errStr == "signature is invalid" ||
			errStr == "token has invalid claims: signature is invalid" ||
			errStr == "token signature is invalid: signature is invalid" {
			return uuid.Nil, ErrInvalidToken
		}
		return uuid.Nil, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	// Check if token is expired (double-check)
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return uuid.Nil, ErrExpiredToken
	}

	// Parse user ID from subject
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	return userID, nil
}

// CreateAccessToken generates a JWT token for a user
func CreateAccessToken(userID uuid.UUID) (string, error) {
	// Use a reasonable default expiration time (1 hour)
	tokenSecret := "default-secret-key" // In production, this should come from environment
	return MakeJWT(userID, tokenSecret, time.Hour)
}
