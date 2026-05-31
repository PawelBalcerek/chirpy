package auth

import (
	"errors"
	"net/http"
	"strings"
)

const (
	apiKeyPrefix = "ApiKey "
)

var (
	ErrMissingApiKey = errors.New("missing api key")
	ErrInvalidApiKey = errors.New("invalid api key")
)

func GetApiKey(headers http.Header) (string, error) {
	authorization := headers.Get("authorization")
	if authorization == "" {
		return "", ErrMissingApiKey
	}

	token, ok := strings.CutPrefix(authorization, apiKeyPrefix)
	if !ok {
		return "", ErrInvalidApiKey
	}

	return token, nil
}
