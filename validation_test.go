package main

import (
	"strings"
	"testing"

	"github.com/kai-xlr/neo_chirpy/internal/auth"
)

func TestValidateChirpBody(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr error
	}{
		{
			name:    "valid chirp",
			body:    "This is a valid chirp",
			wantErr: nil,
		},
		{
			name:    "empty chirp",
			body:    "",
			wantErr: ErrChirpEmpty,
		},
		{
			name:    "whitespace only chirp",
			body:    "   ",
			wantErr: ErrChirpEmpty,
		},
		{
			name:    "chirp too long",
			body:    strings.Repeat("a", MaxChirpLength+1),
			wantErr: ErrChirpTooLong,
		},
		{
			name:    "chirp at max length",
			body:    strings.Repeat("a", MaxChirpLength),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateChirpBody(tt.body)
			if err != tt.wantErr {
				t.Errorf("ValidateChirpBody() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr error
	}{
		{
			name:    "valid email",
			email:   "user@example.com",
			wantErr: nil,
		},
		{
			name:    "empty email",
			email:   "",
			wantErr: ErrEmailEmpty,
		},
		{
			name:    "whitespace only email",
			email:   "   ",
			wantErr: ErrEmailEmpty,
		},
		{
			name:    "email without @",
			email:   "userexample.com",
			wantErr: ErrEmailInvalid,
		},
		{
			name:    "email without domain",
			email:   "user@",
			wantErr: ErrEmailInvalid,
		},
		{
			name:    "email without dot",
			email:   "user@example",
			wantErr: ErrEmailInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if err != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUserUpdateRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     userUpdateRequest
		wantErr error
	}{
		{
			name: "valid update request",
			req: userUpdateRequest{
				Email:    "newemail@example.com",
				Password: "newpassword123",
			},
			wantErr: nil,
		},
		{
			name: "invalid email",
			req: userUpdateRequest{
				Email:    "invalid-email",
				Password: "newpassword123",
			},
			wantErr: ErrEmailInvalid,
		},
		{
			name: "empty email",
			req: userUpdateRequest{
				Email:    "",
				Password: "newpassword123",
			},
			wantErr: ErrEmailEmpty,
		},
		{
			name: "empty password",
			req: userUpdateRequest{
				Email:    "newemail@example.com",
				Password: "",
			},
			wantErr: auth.ErrPasswordEmpty,
		},
		{
			name: "whitespace only email",
			req: userUpdateRequest{
				Email:    "   ",
				Password: "newpassword123",
			},
			wantErr: ErrEmailEmpty,
		},
		{
			name: "whitespace only password",
			req: userUpdateRequest{
				Email:    "newemail@example.com",
				Password: "   ",
			},
			wantErr: auth.ErrPasswordEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUserUpdateRequest(tt.req)
			if err != tt.wantErr {
				t.Errorf("validateUserUpdateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPathMatching(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		prefix string
		want   bool
	}{
		{
			name:   "valid chirp ID path",
			path:   "/api/chirps/123e4567-e89b-12d3-a456-426614174000",
			prefix: "/api/chirps/",
			want:   true,
		},
		{
			name:   "exact match (should be false)",
			path:   "/api/chirps/",
			prefix: "/api/chirps/",
			want:   false,
		},
		{
			name:   "different path",
			path:   "/api/users/123",
			prefix: "/api/chirps/",
			want:   false,
		},
		{
			name:   "shorter than prefix",
			path:   "/api/chirp",
			prefix: "/api/chirps/",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pathMatch(tt.path, tt.prefix)
			if result != tt.want {
				t.Errorf("pathMatch() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestExtractIDFromPath(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		prefix string
		want   string
	}{
		{
			name:   "extract valid chirp ID",
			path:   "/api/chirps/123e4567-e89b-12d3-a456-426614174000",
			prefix: "/api/chirps/",
			want:   "123e4567-e89b-12d3-a456-426614174000",
		},
		{
			name:   "empty result when path equals prefix",
			path:   "/api/chirps/",
			prefix: "/api/chirps/",
			want:   "",
		},
		{
			name:   "empty result when path shorter than prefix",
			path:   "/api/chirp",
			prefix: "/api/chirps/",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractIDFromPath(tt.path, tt.prefix)
			if result != tt.want {
				t.Errorf("extractIDFromPath() = %v, want %v", result, tt.want)
			}
		})
	}
}
