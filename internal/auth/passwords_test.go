package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test-secret-key"
	expiresIn := time.Hour

	token, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT() error = %v", err)
	}

	if token == "" {
		t.Fatal("MakeJWT() returned empty token")
	}
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test-secret-key"
	expiresIn := time.Hour

	// Create a valid token
	token, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT() error = %v", err)
	}

	// Validate the token
	validatedUserID, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Fatalf("ValidateJWT() error = %v", err)
	}

	if validatedUserID != userID {
		t.Errorf("ValidateJWT() userID = %v, want %v", validatedUserID, userID)
	}
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test-secret-key"
	expiresIn := time.Millisecond // Very short expiration

	// Create a token that expires immediately
	token, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT() error = %v", err)
	}

	// Wait for token to expire
	time.Sleep(time.Millisecond * 10)

	// Try to validate the expired token
	_, err = ValidateJWT(token, tokenSecret)
	if err == nil {
		t.Fatal("ValidateJWT() should have returned an error for expired token")
	}

	// Check that it's our expired token error
	if err != ErrExpiredToken {
		t.Errorf("ValidateJWT() error = %v, want %v", err, ErrExpiredToken)
	}
}

func TestValidateJWT_WrongSecret(t *testing.T) {
	userID := uuid.New()
	correctSecret := "correct-secret"
	wrongSecret := "wrong-secret"
	expiresIn := time.Hour

	// Create a token with the correct secret
	token, err := MakeJWT(userID, correctSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT() error = %v", err)
	}

	// Try to validate with the wrong secret
	_, err = ValidateJWT(token, wrongSecret)
	if err == nil {
		t.Fatal("ValidateJWT() should have returned an error for wrong secret")
	}

	// Check that it's our invalid token error
	if err != ErrInvalidToken {
		t.Errorf("ValidateJWT() error = %v, want %v", err, ErrInvalidToken)
	}
}

func TestValidateJWT_InvalidToken(t *testing.T) {
	tokenSecret := "test-secret-key"

	// Test with completely invalid token
	testCases := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"malformed token", "not.a.valid.jwt"},
		{"random string", "random-string-token"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidateJWT(tc.token, tokenSecret)
			if err == nil {
				t.Errorf("ValidateJWT() should have returned an error for %s", tc.name)
			}
		})
	}
}

func TestValidateJWT_ClaimsExtraction(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test-secret-key"
	expiresIn := time.Hour

	// Create a token
	token, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT() error = %v", err)
	}

	// Parse the token directly to verify claims structure
	parsedToken, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	claims, ok := parsedToken.Claims.(*jwt.RegisteredClaims)
	if !ok {
		t.Fatal("Failed to extract claims")
	}

	// Verify issuer
	if claims.Issuer != "chirpy" {
		t.Errorf("Issuer = %v, want %v", claims.Issuer, "chirpy")
	}

	// Verify subject
	if claims.Subject != userID.String() {
		t.Errorf("Subject = %v, want %v", claims.Subject, userID.String())
	}

	// Verify issued at is set
	if claims.IssuedAt == nil {
		t.Error("IssuedAt should not be nil")
	}

	// Verify expires at is set
	if claims.ExpiresAt == nil {
		t.Error("ExpiresAt should not be nil")
	}
}

func TestCreateAccessToken_Integration(t *testing.T) {
	userID := uuid.New()

	// Test that CreateAccessToken still works with the new JWT implementation
	token, err := CreateAccessToken(userID)
	if err != nil {
		t.Fatalf("CreateAccessToken() error = %v", err)
	}

	if token == "" {
		t.Fatal("CreateAccessToken() returned empty token")
	}

	// Validate with the default secret
	validatedUserID, err := ValidateJWT(token, "default-secret-key")
	if err != nil {
		t.Fatalf("ValidateJWT() error = %v", err)
	}

	if validatedUserID != userID {
		t.Errorf("CreateAccessToken() validation failed, got %v, want %v", validatedUserID, userID)
	}
}
