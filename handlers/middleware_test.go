package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/PawelBalcerek/chirpy/handlers"
	"github.com/PawelBalcerek/chirpy/internal/auth"
	"github.com/google/uuid"
)

func TestRequireJWT(t *testing.T) {
	tokenSecret := auth.TokenSecret("my-ultra-secure-test-jwt-secret-key")
	userID := uuid.New()

	token, err := auth.MakeJWT(userID, tokenSecret, time.Hour)
	if err != nil {
		t.Fatalf("failed to make jwt: %v", err)
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID, ok := handlers.GetUserID(r.Context())
		if !ok {
			t.Error("expected userID to be present in context")
		}
		if gotUserID != userID {
			t.Errorf("expected userID %s, got %s", userID, gotUserID)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := handlers.RequireJWT(tokenSecret)(nextHandler)

	t.Run("Valid Token", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/chirps", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rr.Code)
		}
	})

	t.Run("Missing Authorization Header", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/chirps", nil)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rr.Code)
		}
	})

	t.Run("Invalid Token Signature", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/chirps", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()

		badMiddleware := handlers.RequireJWT("wrong-secret")(nextHandler)
		badMiddleware.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rr.Code)
		}
	})
}

func TestRequireApiKey(t *testing.T) {
	expectedKey := "test-api-key"

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey, ok := handlers.GetApiKey(r.Context())
		if !ok {
			t.Error("expected apiKey to be present in context")
		}
		if gotKey != expectedKey {
			t.Errorf("expected apiKey %s, got %s", expectedKey, gotKey)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	middleware := handlers.RequireApiKey(expectedKey)(nextHandler)

	t.Run("Valid API Key", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/polka/webhooks", nil)
		req.Header.Set("Authorization", "ApiKey "+expectedKey)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", rr.Code)
		}
	})

	t.Run("Invalid API Key", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/polka/webhooks", nil)
		req.Header.Set("Authorization", "ApiKey wrong-key")
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", rr.Code)
		}
	})
}
