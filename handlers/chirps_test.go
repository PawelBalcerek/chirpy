package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/PawelBalcerek/chirpy/internal/auth"
	"github.com/PawelBalcerek/chirpy/internal/database"
	"github.com/google/uuid"
)

type fakeChirpStore struct {
	chirps []database.Chirp
}

func (f *fakeChirpStore) CreateChirp(_ context.Context, arg database.CreateChirpParams) (database.Chirp, error) {
	chirp := database.Chirp{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body:      arg.Body,
		UserID:    arg.UserID,
	}
	f.chirps = append(f.chirps, chirp)
	return chirp, nil
}

func (f *fakeChirpStore) GetChirp(_ context.Context, id uuid.UUID) (database.Chirp, error) {
	for _, c := range f.chirps {
		if c.ID == id {
			return c, nil
		}
	}
	return database.Chirp{}, sql.ErrNoRows
}

func (f *fakeChirpStore) GetChirps(_ context.Context, _ uuid.UUID) ([]database.Chirp, error) {
	return f.chirps, nil
}

func (f *fakeChirpStore) DeleteChirp(_ context.Context, id uuid.UUID) error {
	for i, c := range f.chirps {
		if c.ID == id {
			f.chirps = append(f.chirps[:i], f.chirps[i+1:]...)
			return nil
		}
	}
	return nil
}

func makeTestJWT(t *testing.T, userID uuid.UUID, secret string) string {
	t.Helper()
	tok, err := auth.MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("makeTestJWT: %v", err)
	}
	return tok
}

func TestCreateChirpHandler_profanityFiltered(t *testing.T) {
	store := &fakeChirpStore{}
	secret := "test-secret"
	userID := uuid.New()

	handler := &CreateChirpHandler{
		DbQueries: store,
		JWTSecret: secret,
	}

	body, _ := json.Marshal(map[string]string{"body": "This is a kerfuffle situation"})
	req := httptest.NewRequest(http.MethodPost, "/api/chirps", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeTestJWT(t, userID, secret))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp chirpResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	want := "This is a **** situation"
	if resp.Body != want {
		t.Errorf("expected body %q, got %q", want, resp.Body)
	}
}

func TestCreateChirpHandler_rejectsOver140Chars(t *testing.T) {
	store := &fakeChirpStore{}
	secret := "test-secret"
	userID := uuid.New()

	handler := &CreateChirpHandler{
		DbQueries: store,
		JWTSecret: secret,
	}

	longBody := make([]byte, 141)
	for i := range longBody {
		longBody[i] = 'a'
	}
	body, _ := json.Marshal(map[string]string{"body": string(longBody)})
	req := httptest.NewRequest(http.MethodPost, "/api/chirps", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeTestJWT(t, userID, secret))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestCreateChirpHandler_rejectsUnauthenticated(t *testing.T) {
	store := &fakeChirpStore{}
	handler := &CreateChirpHandler{
		DbQueries: store,
		JWTSecret: "test-secret",
	}

	body, _ := json.Marshal(map[string]string{"body": "hello"})
	req := httptest.NewRequest(http.MethodPost, "/api/chirps", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}
