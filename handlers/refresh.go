package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/PawelBalcerek/chirpy/internal/auth"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type RefreshHandler struct {
	DbQueries TokenStore
	JWTSecret string
}

func (h RefreshHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		caser := cases.Title(language.English, cases.NoLower)
		handleError(err, caser.String(err.Error()), w, http.StatusUnauthorized)
		return
	}

	refreshToken, err := h.DbQueries.GetRefreshToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON("Unknown refresh token", w, http.StatusUnauthorized)
			return
		}
		handleError(err, "Failed to obtain refresh token", w, http.StatusInternalServerError)
		return
	}

	if refreshToken.RevokedAt.Valid {
		writeJSON("Refresh token has been revoked", w, http.StatusUnauthorized)
		return
	}

	if !refreshToken.ExpiresAt.After(time.Now()) {
		writeJSON("Expired refresh token", w, http.StatusUnauthorized)
		return
	}

	jwt, err := auth.MakeJWT(refreshToken.UserID, h.JWTSecret, jwtExpiresIn)
	if err != nil {
		handleError(err, "Failed to make JWT", w, http.StatusInternalServerError)
		return
	}

	type refreshResponse struct {
		Token string `json:"token"`
	}
	writeJSON(refreshResponse{Token: jwt}, w, http.StatusOK)
}
