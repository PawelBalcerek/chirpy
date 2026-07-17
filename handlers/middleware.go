package handlers

import (
	"context"
	"net/http"

	"github.com/PawelBalcerek/chirpy/internal/auth"
	"github.com/google/uuid"
)

type contextKey string

const (
	userIDContextKey contextKey = "userID"
	apiKeyContextKey contextKey = "apiKey"
)

func RequireJWT(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := auth.GetBearerToken(r.Header)
			if err != nil {
				handleAuthorizationError(err, w)
				return
			}
			userID, err := auth.ValidateJWT(token, jwtSecret)
			if err != nil {
				handleError(err, "Invalid token", w, http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireApiKey(expectedKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey, err := auth.GetApiKey(r.Header)
			if err != nil {
				handleAuthorizationError(err, w)
				return
			}
			if apiKey != expectedKey {
				handleError(nil, "Invalid api key", w, http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), apiKeyContextKey, apiKey)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDContextKey).(uuid.UUID)
	return userID, ok
}

func GetApiKey(ctx context.Context) (string, bool) {
	apiKey, ok := ctx.Value(apiKeyContextKey).(string)
	return apiKey, ok
}
