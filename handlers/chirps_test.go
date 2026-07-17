package handlers_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/PawelBalcerek/chirpy/handlers"
	"github.com/PawelBalcerek/chirpy/internal/database"
	"github.com/google/uuid"
)

func TestChirpCreate_HappyPath(t *testing.T) {
	userID := uuid.New()
	chirpID := uuid.New()
	now := time.Now()

	store := &fakeChirpStore{
		CreateChirpFunc: func(_ context.Context, arg database.CreateChirpParams) (database.Chirp, error) {
			return database.Chirp{
				ID:        chirpID,
				CreatedAt: now,
				UpdatedAt: now,
				Body:      arg.Body,
				UserID:    arg.UserID,
			}, nil
		},
	}
	ctrl := &handlers.ChirpController{ChirpStore: store}

	body := `{"body":"hello world"}`
	req := httptest.NewRequest(http.MethodPost, "/api/chirps", strings.NewReader(body))
	req = handlers.WithUserIDContext(req, userID)
	rr := httptest.NewRecorder()

	ctrl.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp["body"] != "hello world" {
		t.Errorf("expected body 'hello world', got %v", resp["body"])
	}
}

func TestChirpCreate_ProfanityFiltered(t *testing.T) {
	userID := uuid.New()
	store := &fakeChirpStore{
		CreateChirpFunc: func(_ context.Context, arg database.CreateChirpParams) (database.Chirp, error) {
			return database.Chirp{Body: arg.Body, UserID: arg.UserID}, nil
		},
	}
	ctrl := &handlers.ChirpController{ChirpStore: store}

	body := `{"body":"what the kerfuffle is this"}`
	req := httptest.NewRequest(http.MethodPost, "/api/chirps", strings.NewReader(body))
	req = handlers.WithUserIDContext(req, userID)
	rr := httptest.NewRecorder()

	ctrl.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
	var resp map[string]any
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["body"] != "what the **** is this" {
		t.Errorf("expected profanity replaced, got %v", resp["body"])
	}
}

func TestChirpCreate_BodyTooLong(t *testing.T) {
	userID := uuid.New()
	ctrl := &handlers.ChirpController{ChirpStore: &fakeChirpStore{}}

	long := strings.Repeat("a", 141)
	body := fmt.Sprintf(`{"body":"%s"}`, long)
	req := httptest.NewRequest(http.MethodPost, "/api/chirps", strings.NewReader(body))
	req = handlers.WithUserIDContext(req, userID)
	rr := httptest.NewRecorder()

	ctrl.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestChirpCreate_EmptyBody(t *testing.T) {
	userID := uuid.New()
	ctrl := &handlers.ChirpController{ChirpStore: &fakeChirpStore{}}

	body := `{"body":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/chirps", strings.NewReader(body))
	req = handlers.WithUserIDContext(req, userID)
	rr := httptest.NewRecorder()

	ctrl.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestChirpCreate_StoreError(t *testing.T) {
	userID := uuid.New()
	store := &fakeChirpStore{
		CreateChirpFunc: func(_ context.Context, _ database.CreateChirpParams) (database.Chirp, error) {
			return database.Chirp{}, errors.New("db error")
		},
	}
	ctrl := &handlers.ChirpController{ChirpStore: store}

	body := `{"body":"fine body"}`
	req := httptest.NewRequest(http.MethodPost, "/api/chirps", strings.NewReader(body))
	req = handlers.WithUserIDContext(req, userID)
	rr := httptest.NewRecorder()

	ctrl.Create(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestChirpCreate_NoUserIDInContext(t *testing.T) {
	ctrl := &handlers.ChirpController{ChirpStore: &fakeChirpStore{}}

	req := httptest.NewRequest(http.MethodPost, "/api/chirps", strings.NewReader(`{"body":"hi"}`))
	rr := httptest.NewRecorder()

	ctrl.Create(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestChirpGet_HappyPath(t *testing.T) {
	chirpID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	store := &fakeChirpStore{
		GetChirpFunc: func(_ context.Context, id uuid.UUID) (database.Chirp, error) {
			if id == chirpID {
				return database.Chirp{ID: chirpID, CreatedAt: now, UpdatedAt: now, Body: "test", UserID: userID}, nil
			}
			return database.Chirp{}, sql.ErrNoRows
		},
	}
	ctrl := &handlers.ChirpController{ChirpStore: store}

	req := httptest.NewRequest(http.MethodGet, "/api/chirps/"+chirpID.String(), nil)
	req.SetPathValue("id", chirpID.String())
	rr := httptest.NewRecorder()

	ctrl.Get(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestChirpGet_NotFound(t *testing.T) {
	store := &fakeChirpStore{
		GetChirpFunc: func(_ context.Context, _ uuid.UUID) (database.Chirp, error) {
			return database.Chirp{}, sql.ErrNoRows
		},
	}
	ctrl := &handlers.ChirpController{ChirpStore: store}

	id := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/chirps/"+id.String(), nil)
	req.SetPathValue("id", id.String())
	rr := httptest.NewRecorder()

	ctrl.Get(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestChirpGet_InvalidID(t *testing.T) {
	ctrl := &handlers.ChirpController{ChirpStore: &fakeChirpStore{}}

	req := httptest.NewRequest(http.MethodGet, "/api/chirps/not-a-uuid", nil)
	req.SetPathValue("id", "not-a-uuid")
	rr := httptest.NewRecorder()

	ctrl.Get(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestChirpGet_StoreError(t *testing.T) {
	store := &fakeChirpStore{
		GetChirpFunc: func(_ context.Context, _ uuid.UUID) (database.Chirp, error) {
			return database.Chirp{}, errors.New("db error")
		},
	}
	ctrl := &handlers.ChirpController{ChirpStore: store}

	id := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/chirps/"+id.String(), nil)
	req.SetPathValue("id", id.String())
	rr := httptest.NewRecorder()

	ctrl.Get(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestChirpList_All(t *testing.T) {
	now := time.Now()
	chirps := []database.Chirp{
		{ID: uuid.New(), CreatedAt: now, UpdatedAt: now, Body: "first", UserID: uuid.New()},
		{ID: uuid.New(), CreatedAt: now.Add(time.Second), UpdatedAt: now, Body: "second", UserID: uuid.New()},
	}
	store := &fakeChirpStore{
		GetChirpsFunc: func(_ context.Context, _ uuid.UUID) ([]database.Chirp, error) {
			return chirps, nil
		},
	}
	ctrl := &handlers.ChirpController{ChirpStore: store}

	req := httptest.NewRequest(http.MethodGet, "/api/chirps", nil)
	rr := httptest.NewRecorder()

	ctrl.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	var resp []map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("expected 2 chirps, got %d", len(resp))
	}
}

func TestChirpList_SortDesc(t *testing.T) {
	now := time.Now()
	userID := uuid.New()
	chirps := []database.Chirp{
		{ID: uuid.New(), CreatedAt: now, UpdatedAt: now, Body: "older", UserID: userID},
		{ID: uuid.New(), CreatedAt: now.Add(time.Second), UpdatedAt: now, Body: "newer", UserID: userID},
	}
	store := &fakeChirpStore{
		GetChirpsFunc: func(_ context.Context, _ uuid.UUID) ([]database.Chirp, error) {
			return chirps, nil
		},
	}
	ctrl := &handlers.ChirpController{ChirpStore: store}

	req := httptest.NewRequest(http.MethodGet, "/api/chirps?sort=desc", nil)
	rr := httptest.NewRecorder()

	ctrl.List(rr, req)

	var resp []map[string]any
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if len(resp) != 2 {
		t.Fatalf("expected 2 chirps, got %d", len(resp))
	}
	if resp[0]["body"] != "newer" {
		t.Errorf("expected 'newer' first in desc order, got %v", resp[0]["body"])
	}
}

func TestChirpList_InvalidAuthorID(t *testing.T) {
	ctrl := &handlers.ChirpController{ChirpStore: &fakeChirpStore{}}

	req := httptest.NewRequest(http.MethodGet, "/api/chirps?author_id=not-a-uuid", nil)
	rr := httptest.NewRecorder()

	ctrl.List(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestChirpList_StoreError(t *testing.T) {
	store := &fakeChirpStore{
		GetChirpsFunc: func(_ context.Context, _ uuid.UUID) ([]database.Chirp, error) {
			return nil, errors.New("db error")
		},
	}
	ctrl := &handlers.ChirpController{ChirpStore: store}

	req := httptest.NewRequest(http.MethodGet, "/api/chirps", nil)
	rr := httptest.NewRecorder()

	ctrl.List(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestChirpDelete_HappyPath(t *testing.T) {
	chirpID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	store := &fakeChirpStore{
		GetChirpFunc: func(_ context.Context, _ uuid.UUID) (database.Chirp, error) {
			return database.Chirp{ID: chirpID, CreatedAt: now, UpdatedAt: now, Body: "bye", UserID: userID}, nil
		},
		DeleteChirpFunc: func(_ context.Context, _ uuid.UUID) error {
			return nil
		},
	}
	ctrl := &handlers.ChirpController{ChirpStore: store}

	req := httptest.NewRequest(http.MethodDelete, "/api/chirps/"+chirpID.String(), nil)
	req.SetPathValue("id", chirpID.String())
	req = handlers.WithUserIDContext(req, userID)
	rr := httptest.NewRecorder()

	ctrl.Delete(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

func TestChirpDelete_Forbidden(t *testing.T) {
	chirpID := uuid.New()
	ownerID := uuid.New()
	callerID := uuid.New()
	now := time.Now()

	store := &fakeChirpStore{
		GetChirpFunc: func(_ context.Context, _ uuid.UUID) (database.Chirp, error) {
			return database.Chirp{ID: chirpID, CreatedAt: now, UpdatedAt: now, Body: "owned", UserID: ownerID}, nil
		},
	}
	ctrl := &handlers.ChirpController{ChirpStore: store}

	req := httptest.NewRequest(http.MethodDelete, "/api/chirps/"+chirpID.String(), nil)
	req.SetPathValue("id", chirpID.String())
	req = handlers.WithUserIDContext(req, callerID)
	rr := httptest.NewRecorder()

	ctrl.Delete(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestChirpDelete_NotFound(t *testing.T) {
	store := &fakeChirpStore{
		GetChirpFunc: func(_ context.Context, _ uuid.UUID) (database.Chirp, error) {
			return database.Chirp{}, sql.ErrNoRows
		},
	}
	ctrl := &handlers.ChirpController{ChirpStore: store}

	id := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/chirps/"+id.String(), nil)
	req.SetPathValue("id", id.String())
	req = handlers.WithUserIDContext(req, uuid.New())
	rr := httptest.NewRecorder()

	ctrl.Delete(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestChirpDelete_NoUserIDInContext(t *testing.T) {
	ctrl := &handlers.ChirpController{ChirpStore: &fakeChirpStore{}}

	id := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/chirps/"+id.String(), nil)
	req.SetPathValue("id", id.String())
	rr := httptest.NewRecorder()

	ctrl.Delete(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestChirpDelete_InvalidID(t *testing.T) {
	ctrl := &handlers.ChirpController{ChirpStore: &fakeChirpStore{}}

	req := httptest.NewRequest(http.MethodDelete, "/api/chirps/not-a-uuid", nil)
	req.SetPathValue("id", "not-a-uuid")
	req = handlers.WithUserIDContext(req, uuid.New())
	rr := httptest.NewRecorder()

	ctrl.Delete(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestChirpDelete_DeleteStoreError(t *testing.T) {
	chirpID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	store := &fakeChirpStore{
		GetChirpFunc: func(_ context.Context, _ uuid.UUID) (database.Chirp, error) {
			return database.Chirp{ID: chirpID, CreatedAt: now, UpdatedAt: now, Body: "bye", UserID: userID}, nil
		},
		DeleteChirpFunc: func(_ context.Context, _ uuid.UUID) error {
			return errors.New("db error")
		},
	}
	ctrl := &handlers.ChirpController{ChirpStore: store}

	req := httptest.NewRequest(http.MethodDelete, "/api/chirps/"+chirpID.String(), nil)
	req.SetPathValue("id", chirpID.String())
	req = handlers.WithUserIDContext(req, userID)
	rr := httptest.NewRecorder()

	ctrl.Delete(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}
