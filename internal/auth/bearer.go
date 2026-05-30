package auth

import (
	"errors"
	"net/http"
	"strings"
)

const (
	bearerTokenPrefix = "Bearer "
)

var (
	ErrMissingBearerToken = errors.New("missing bearer token")
	ErrInvalidBearerToken = errors.New("invalid bearer token")
)

func GetBearerToken(headers http.Header) (string, error) {
	authorization := headers.Get("authorization")
	if authorization == "" {
		return "", ErrMissingBearerToken
	}

	token, ok := strings.CutPrefix(authorization, bearerTokenPrefix)
	if !ok {
		return "", ErrInvalidBearerToken
	}

	return token, nil
}
