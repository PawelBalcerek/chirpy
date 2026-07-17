package chirp_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/PawelBalcerek/chirpy/internal/chirp"
)

func TestNewBody_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid simple text",
			input:    "Hello world!",
			expected: "Hello world!",
		},
		{
			name:     "filters profane word",
			input:    "This is a kerfuffle",
			expected: "This is a ****",
		},
		{
			name:     "filters multiple profane words case insensitively",
			input:    "What a KerFuffle and SHARBERT indeed",
			expected: "What a **** and **** indeed",
		},
		{
			name:     "ignores sub-words containing profanities",
			input:    "Fornaxian is not exactly fornax",
			expected: "Fornaxian is not exactly ****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := chirp.NewBody(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if body.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, body.String())
			}
		})
	}
}

func TestNewBody_Errors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError error
	}{
		{
			name:          "empty body",
			input:         "",
			expectedError: chirp.ErrBodyEmpty,
		},
		{
			name:          "too long body",
			input:         strings.Repeat("a", 141),
			expectedError: chirp.ErrBodyTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := chirp.NewBody(tt.input)
			if !errors.Is(err, tt.expectedError) {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}
		})
	}
}
