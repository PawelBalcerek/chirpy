package auth_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/PawelBalcerek/chirpy/internal/auth"
)

const (
	token = "TOKEN"
)

func TestGetBearerToken_ReturnsToken(t *testing.T) {
	headers := http.Header{}
	headers.Add("authorization", fmt.Sprintf("Bearer %s", token))
	resultToken, err := auth.GetBearerToken(headers)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if resultToken != token {
		t.Fatalf("expected %s, got: %s", token, resultToken)
	}
}

func TestGetBearerToken_ReturnsMissingBearerTokenError(t *testing.T) {
	_, err := auth.GetBearerToken(http.Header{})
	if !errors.Is(err, auth.ErrMissingBearerToken) {
		t.Fatalf("expected missing bearer token error, got: %v", err)
	}
}

func TestGetBearerToken_ReturnsInvalidBearerTokenError(t *testing.T) {
	headers := http.Header{}
	headers.Add("authorization", fmt.Sprintf("bearer %s", token))
	_, err := auth.GetBearerToken(headers)
	if !errors.Is(err, auth.ErrInvalidBearerToken) {
		t.Fatalf("expected invalid bearer token error, got: %v", err)
	}
}
