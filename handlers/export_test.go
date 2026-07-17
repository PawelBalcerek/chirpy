package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

func WithUserIDContext(r *http.Request, id uuid.UUID) *http.Request {
	ctx := context.WithValue(r.Context(), userIDContextKey, id)
	return r.WithContext(ctx)
}
