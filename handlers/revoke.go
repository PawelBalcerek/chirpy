package handlers

import (
	"net/http"

	"github.com/PawelBalcerek/chirpy/internal/auth"
)

type RevokeHandler struct {
	DbQueries TokenStore
}

func (h RevokeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		handleAuthorizationError(err, w)
		return
	}

	if err := h.DbQueries.RevokeRefreshToken(r.Context(), token); err != nil {
		handleError(err, "Failed to revoke refresh token", w, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
