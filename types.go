package main

import (
	"time"

	"github.com/google/uuid"
)

// Chirp types
type chirpRequest struct {
	Body string `json:"body"`
}

type chirpResponse struct {
	CleanedBody string `json:"cleaned_body"`
}

type chirpCreateRequest struct {
	Body string `json:"body"`
}

type chirpCreateResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `json:"user_id"`
	Body      string    `json:"body"`
}

// User types
type userRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type userResponse struct {
	User
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

type refreshResponse struct {
	Token string `json:"token"`
}

type userUpdateRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
