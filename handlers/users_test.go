package handlers_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/PawelBalcerek/chirpy/handlers"
	"github.com/PawelBalcerek/chirpy/internal/auth"
	"github.com/PawelBalcerek/chirpy/internal/database"
	"github.com/google/uuid"
)

func TestUserCreate_HappyPath(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	store := &fakeUserStore{
		CreateUserFunc: func(_ context.Context, arg database.CreateUserParams) (database.User, error) {
			return database.User{
				ID:        userID,
				CreatedAt: now,
				UpdatedAt: now,
				Email:     arg.Email,
			}, nil
		},
	}
	ctrl := &handlers.UserController{UserStore: store, TokenStore: &fakeTokenStore{}, JWTSecret: "secret"}

	body := `{"email":"user@example.com","password":"hunter2"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))
	rr := httptest.NewRecorder()

	ctrl.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp["email"] != "user@example.com" {
		t.Errorf("expected email 'user@example.com', got %v", resp["email"])
	}
}

func TestUserCreate_InvalidJSON(t *testing.T) {
	ctrl := &handlers.UserController{UserStore: &fakeUserStore{}, TokenStore: &fakeTokenStore{}, JWTSecret: "secret"}

	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader("not-json"))
	rr := httptest.NewRecorder()

	ctrl.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestUserCreate_StoreError(t *testing.T) {
	store := &fakeUserStore{
		CreateUserFunc: func(_ context.Context, _ database.CreateUserParams) (database.User, error) {
			return database.User{}, errors.New("db error")
		},
	}
	ctrl := &handlers.UserController{UserStore: store, TokenStore: &fakeTokenStore{}, JWTSecret: "secret"}

	body := `{"email":"user@example.com","password":"hunter2"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))
	rr := httptest.NewRecorder()

	ctrl.Create(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestUserUpdate_HappyPath(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	store := &fakeUserStore{
		UpdateUserFunc: func(_ context.Context, arg database.UpdateUserParams) (database.User, error) {
			return database.User{
				ID:        arg.ID,
				CreatedAt: now,
				UpdatedAt: now,
				Email:     arg.Email,
			}, nil
		},
	}
	ctrl := &handlers.UserController{UserStore: store, TokenStore: &fakeTokenStore{}, JWTSecret: "secret"}

	body := `{"email":"new@example.com","password":"newpass"}`
	req := httptest.NewRequest(http.MethodPut, "/api/users", strings.NewReader(body))
	req = handlers.WithUserIDContext(req, userID)
	rr := httptest.NewRecorder()

	ctrl.Update(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp map[string]any
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["email"] != "new@example.com" {
		t.Errorf("expected updated email, got %v", resp["email"])
	}
}

func TestUserUpdate_NoUserIDInContext(t *testing.T) {
	ctrl := &handlers.UserController{UserStore: &fakeUserStore{}, TokenStore: &fakeTokenStore{}, JWTSecret: "secret"}

	body := `{"email":"new@example.com","password":"newpass"}`
	req := httptest.NewRequest(http.MethodPut, "/api/users", strings.NewReader(body))
	rr := httptest.NewRecorder()

	ctrl.Update(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestUserUpdate_StoreError(t *testing.T) {
	store := &fakeUserStore{
		UpdateUserFunc: func(_ context.Context, _ database.UpdateUserParams) (database.User, error) {
			return database.User{}, errors.New("db error")
		},
	}
	ctrl := &handlers.UserController{UserStore: store, TokenStore: &fakeTokenStore{}, JWTSecret: "secret"}

	body := `{"email":"new@example.com","password":"newpass"}`
	req := httptest.NewRequest(http.MethodPut, "/api/users", strings.NewReader(body))
	req = handlers.WithUserIDContext(req, uuid.New())
	rr := httptest.NewRecorder()

	ctrl.Update(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestUserLogin_HappyPath(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	password := "correct-horse-battery-staple"
	hashed, _ := auth.HashPassword(password)

	userStore := &fakeUserStore{
		GetUserFunc: func(_ context.Context, email string) (database.User, error) {
			return database.User{
				ID:             userID,
				CreatedAt:      now,
				UpdatedAt:      now,
				Email:          email,
				HashedPassword: hashed,
			}, nil
		},
	}
	tokenStore := &fakeTokenStore{
		CreateRefreshTokenFunc: func(_ context.Context, _ database.CreateRefreshTokenParams) (database.RefreshToken, error) {
			return database.RefreshToken{}, nil
		},
	}
	ctrl := &handlers.UserController{UserStore: userStore, TokenStore: tokenStore, JWTSecret: "secret"}

	body := `{"email":"user@example.com","password":"correct-horse-battery-staple"}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))
	rr := httptest.NewRecorder()

	ctrl.Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["token"] == nil || resp["token"] == "" {
		t.Error("expected a JWT token in response")
	}
	if resp["refresh_token"] == nil || resp["refresh_token"] == "" {
		t.Error("expected a refresh_token in response")
	}
}

func TestUserLogin_WrongPassword(t *testing.T) {
	password := "correct-password"
	hashed, _ := auth.HashPassword(password)

	userStore := &fakeUserStore{
		GetUserFunc: func(_ context.Context, _ string) (database.User, error) {
			return database.User{HashedPassword: hashed}, nil
		},
	}
	ctrl := &handlers.UserController{UserStore: userStore, TokenStore: &fakeTokenStore{}, JWTSecret: "secret"}

	body := `{"email":"user@example.com","password":"wrong-password"}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))
	rr := httptest.NewRecorder()

	ctrl.Login(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestUserLogin_UserNotFound(t *testing.T) {
	userStore := &fakeUserStore{
		GetUserFunc: func(_ context.Context, _ string) (database.User, error) {
			return database.User{}, sql.ErrNoRows
		},
	}
	ctrl := &handlers.UserController{UserStore: userStore, TokenStore: &fakeTokenStore{}, JWTSecret: "secret"}

	body := `{"email":"ghost@example.com","password":"anything"}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))
	rr := httptest.NewRecorder()

	ctrl.Login(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestUserLogin_CreateRefreshTokenError(t *testing.T) {
	password := "mypassword"
	hashed, _ := auth.HashPassword(password)

	userStore := &fakeUserStore{
		GetUserFunc: func(_ context.Context, _ string) (database.User, error) {
			return database.User{ID: uuid.New(), HashedPassword: hashed}, nil
		},
	}
	tokenStore := &fakeTokenStore{
		CreateRefreshTokenFunc: func(_ context.Context, _ database.CreateRefreshTokenParams) (database.RefreshToken, error) {
			return database.RefreshToken{}, errors.New("db error")
		},
	}
	ctrl := &handlers.UserController{UserStore: userStore, TokenStore: tokenStore, JWTSecret: "secret"}

	body := `{"email":"user@example.com","password":"mypassword"}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))
	rr := httptest.NewRecorder()

	ctrl.Login(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestUserRefresh_HappyPath(t *testing.T) {
	userID := uuid.New()
	rawToken := "valid-refresh-token"

	tokenStore := &fakeTokenStore{
		GetRefreshTokenFunc: func(_ context.Context, token string) (database.RefreshToken, error) {
			return database.RefreshToken{
				Token:     token,
				UserID:    userID,
				ExpiresAt: time.Now().Add(time.Hour),
				RevokedAt: sql.NullTime{Valid: false},
			}, nil
		},
	}
	ctrl := &handlers.UserController{UserStore: &fakeUserStore{}, TokenStore: tokenStore, JWTSecret: "secret"}

	req := httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
	req.Header.Set("Authorization", "Bearer "+rawToken)
	rr := httptest.NewRecorder()

	ctrl.Refresh(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["token"] == nil || resp["token"] == "" {
		t.Error("expected a new JWT in response")
	}
}

func TestUserRefresh_RevokedToken(t *testing.T) {
	tokenStore := &fakeTokenStore{
		GetRefreshTokenFunc: func(_ context.Context, _ string) (database.RefreshToken, error) {
			return database.RefreshToken{
				ExpiresAt: time.Now().Add(time.Hour),
				RevokedAt: sql.NullTime{Time: time.Now(), Valid: true},
			}, nil
		},
	}
	ctrl := &handlers.UserController{UserStore: &fakeUserStore{}, TokenStore: tokenStore, JWTSecret: "secret"}

	req := httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
	req.Header.Set("Authorization", "Bearer some-token")
	rr := httptest.NewRecorder()

	ctrl.Refresh(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestUserRefresh_ExpiredToken(t *testing.T) {
	tokenStore := &fakeTokenStore{
		GetRefreshTokenFunc: func(_ context.Context, _ string) (database.RefreshToken, error) {
			return database.RefreshToken{
				ExpiresAt: time.Now().Add(-time.Hour),
				RevokedAt: sql.NullTime{Valid: false},
			}, nil
		},
	}
	ctrl := &handlers.UserController{UserStore: &fakeUserStore{}, TokenStore: tokenStore, JWTSecret: "secret"}

	req := httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
	req.Header.Set("Authorization", "Bearer some-token")
	rr := httptest.NewRecorder()

	ctrl.Refresh(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestUserRefresh_UnknownToken(t *testing.T) {
	tokenStore := &fakeTokenStore{
		GetRefreshTokenFunc: func(_ context.Context, _ string) (database.RefreshToken, error) {
			return database.RefreshToken{}, sql.ErrNoRows
		},
	}
	ctrl := &handlers.UserController{UserStore: &fakeUserStore{}, TokenStore: tokenStore, JWTSecret: "secret"}

	req := httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
	req.Header.Set("Authorization", "Bearer unknown-token")
	rr := httptest.NewRecorder()

	ctrl.Refresh(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestUserRefresh_MissingHeader(t *testing.T) {
	ctrl := &handlers.UserController{UserStore: &fakeUserStore{}, TokenStore: &fakeTokenStore{}, JWTSecret: "secret"}

	req := httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
	rr := httptest.NewRecorder()

	ctrl.Refresh(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestUserRevoke_HappyPath(t *testing.T) {
	tokenStore := &fakeTokenStore{
		RevokeRefreshTokenFunc: func(_ context.Context, _ string) error {
			return nil
		},
	}
	ctrl := &handlers.UserController{UserStore: &fakeUserStore{}, TokenStore: tokenStore, JWTSecret: "secret"}

	req := httptest.NewRequest(http.MethodPost, "/api/revoke", nil)
	req.Header.Set("Authorization", "Bearer some-token")
	rr := httptest.NewRecorder()

	ctrl.Revoke(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

func TestUserRevoke_MissingHeader(t *testing.T) {
	ctrl := &handlers.UserController{UserStore: &fakeUserStore{}, TokenStore: &fakeTokenStore{}, JWTSecret: "secret"}

	req := httptest.NewRequest(http.MethodPost, "/api/revoke", nil)
	rr := httptest.NewRecorder()

	ctrl.Revoke(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestUserRevoke_StoreError(t *testing.T) {
	tokenStore := &fakeTokenStore{
		RevokeRefreshTokenFunc: func(_ context.Context, _ string) error {
			return errors.New("db error")
		},
	}
	ctrl := &handlers.UserController{UserStore: &fakeUserStore{}, TokenStore: tokenStore, JWTSecret: "secret"}

	req := httptest.NewRequest(http.MethodPost, "/api/revoke", nil)
	req.Header.Set("Authorization", "Bearer some-token")
	rr := httptest.NewRecorder()

	ctrl.Revoke(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}
