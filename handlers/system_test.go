package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PawelBalcerek/chirpy/handlers"
	"github.com/PawelBalcerek/chirpy/metrics"
)

func TestHealthCheck(t *testing.T) {
	ctrl := &handlers.SystemController{
		Metrics:   &metrics.Metrics{},
		UserStore: &fakeUserStore{},
		Platform:  "local",
	}

	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	rr := httptest.NewRecorder()

	ctrl.HealthCheck(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "text/plain") {
		t.Errorf("expected text/plain content-type, got %q", ct)
	}
	if body := rr.Body.String(); body != "OK" {
		t.Errorf("expected body 'OK', got %q", body)
	}
}

func TestMetricsReport(t *testing.T) {
	m := &metrics.Metrics{}
	m.FileserverHitsInc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).
		ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/app/", nil))
	m.FileserverHitsInc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).
		ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/app/", nil))

	ctrl := &handlers.SystemController{
		Metrics:   m,
		UserStore: &fakeUserStore{},
		Platform:  "local",
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/metrics", nil)
	rr := httptest.NewRecorder()

	ctrl.MetricsReport(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("expected text/html content-type, got %q", ct)
	}
	if body := rr.Body.String(); !strings.Contains(body, "2") {
		t.Errorf("expected hit count 2 in body, got: %s", body)
	}
}

func TestReset_HappyPath(t *testing.T) {
	m := &metrics.Metrics{}
	store := &fakeUserStore{
		DeleteUsersFunc: func(_ context.Context) error { return nil },
	}
	ctrl := &handlers.SystemController{Metrics: m, UserStore: store, Platform: "local"}

	req := httptest.NewRequest(http.MethodPost, "/admin/reset", nil)
	rr := httptest.NewRecorder()

	ctrl.Reset(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestReset_NonLocalPlatform(t *testing.T) {
	ctrl := &handlers.SystemController{
		Metrics:   &metrics.Metrics{},
		UserStore: &fakeUserStore{},
		Platform:  "production",
	}

	req := httptest.NewRequest(http.MethodPost, "/admin/reset", nil)
	rr := httptest.NewRecorder()

	ctrl.Reset(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestReset_StoreError(t *testing.T) {
	store := &fakeUserStore{
		DeleteUsersFunc: func(_ context.Context) error { return errors.New("db error") },
	}
	ctrl := &handlers.SystemController{
		Metrics:   &metrics.Metrics{},
		UserStore: store,
		Platform:  "local",
	}

	req := httptest.NewRequest(http.MethodPost, "/admin/reset", nil)
	rr := httptest.NewRecorder()

	ctrl.Reset(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}
