package main

import (
	"errors"
	"strings"
)

var (
	ErrChirpTooLong   = errors.New("Chirp is too long")
	ErrChirpEmpty     = errors.New("Chirp cannot be empty")
	ErrEmailInvalid   = errors.New("Invalid email address")
	ErrEmailEmpty     = errors.New("Email cannot be empty")
	ErrUserIDInvalid  = errors.New("Invalid user ID")
)

// ValidateChirpBody validates a chirp body
func ValidateChirpBody(body string) error {
	trimmed := strings.TrimSpace(body)
	
	if trimmed == "" {
		return ErrChirpEmpty
	}
	
	if len(body) > MaxChirpLength {
		return ErrChirpTooLong
	}
	
	return nil
}

// ValidateEmail validates an email address
func ValidateEmail(email string) error {
	trimmed := strings.TrimSpace(email)
	
	if trimmed == "" {
		return ErrEmailEmpty
	}
	
	// Basic email validation (contains @ and .)
	if !strings.Contains(trimmed, "@") || !strings.Contains(trimmed, ".") {
		return ErrEmailInvalid
	}
	
	return nil
}
