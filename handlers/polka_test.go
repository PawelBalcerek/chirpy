package handlers_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PawelBalcerek/chirpy/handlers"
	"github.com/PawelBalcerek/chirpy/internal/database"
	"github.com/google/uuid"
)

func TestPolkaReceiveWebhook_UnknownEvent(t *testing.T) {
	ctrl := &handlers.PolkaController{UserStore: &fakeUserStore{}}

	body := `{"event":"user.downgraded","data":{"user_id":"` + uuid.New().String() + `"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/polka/webhooks", strings.NewReader(body))
	rr := httptest.NewRecorder()

	ctrl.ReceiveWebhook(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 for unrecognised event, got %d", rr.Code)
	}
}

func TestPolkaReceiveWebhook_UserUpgraded_HappyPath(t *testing.T) {
	userID := uuid.New()

	store := &fakeUserStore{
		MakeUserChirpyRedFunc: func(_ context.Context, id uuid.UUID) (database.User, error) {
			if id == userID {
				return database.User{ID: id, IsChirpyRed: true}, nil
			}
			return database.User{}, sql.ErrNoRows
		},
	}
	ctrl := &handlers.PolkaController{UserStore: store}

	body := `{"event":"user.upgraded","data":{"user_id":"` + userID.String() + `"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/polka/webhooks", strings.NewReader(body))
	rr := httptest.NewRecorder()

	ctrl.ReceiveWebhook(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

func TestPolkaReceiveWebhook_UserNotFound(t *testing.T) {
	store := &fakeUserStore{
		MakeUserChirpyRedFunc: func(_ context.Context, _ uuid.UUID) (database.User, error) {
			return database.User{}, sql.ErrNoRows
		},
	}
	ctrl := &handlers.PolkaController{UserStore: store}

	body := `{"event":"user.upgraded","data":{"user_id":"` + uuid.New().String() + `"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/polka/webhooks", strings.NewReader(body))
	rr := httptest.NewRecorder()

	ctrl.ReceiveWebhook(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestPolkaReceiveWebhook_InvalidJSON(t *testing.T) {
	ctrl := &handlers.PolkaController{UserStore: &fakeUserStore{}}

	req := httptest.NewRequest(http.MethodPost, "/api/polka/webhooks", strings.NewReader("not-json"))
	rr := httptest.NewRecorder()

	ctrl.ReceiveWebhook(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestPolkaReceiveWebhook_StoreError(t *testing.T) {
	store := &fakeUserStore{
		MakeUserChirpyRedFunc: func(_ context.Context, _ uuid.UUID) (database.User, error) {
			return database.User{}, errors.New("db error")
		},
	}
	ctrl := &handlers.PolkaController{UserStore: store}

	body := `{"event":"user.upgraded","data":{"user_id":"` + uuid.New().String() + `"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/polka/webhooks", strings.NewReader(body))
	rr := httptest.NewRecorder()

	ctrl.ReceiveWebhook(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}
