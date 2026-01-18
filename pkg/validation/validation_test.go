package validation

import (
	"strings"
	"testing"
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
