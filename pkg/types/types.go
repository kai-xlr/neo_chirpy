package types

import (
	"time"

	"github.com/google/uuid"
)

// Chirp types
type ChirpRequest struct {
	Body string `json:"body"`
}

type ChirpResponse struct {
	CleanedBody string `json:"cleaned_body"`
}

type ChirpCreateRequest struct {
	Body string `json:"body"`
}

type ChirpCreateResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `json:"user_id"`
	Body      string    `json:"body"`
}

// User types
type UserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type User struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

type UserResponse struct {
	User
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

type RefreshResponse struct {
	Token string `json:"token"`
}

type UserUpdateRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Webhook types
type WebhookRequest struct {
	Event string      `json:"event"`
	Data  WebhookData `json:"data"`
}

type WebhookData struct {
	UserID uuid.UUID `json:"user_id"`
}
