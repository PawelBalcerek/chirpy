package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/PawelBalcerek/chirpy/internal/auth"
	"github.com/PawelBalcerek/chirpy/internal/database"
	"github.com/google/uuid"
)

const (
	UnauthorizedUserMsg   = "Incorrect email or password"
	defaultExpiresInHours = 1 * time.Hour
)

type LoginHandler struct {
	DbQueries *database.Queries
	JWTSecret string
}

func (h LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	type loginRequest struct {
		Password         string        `json:"password"`
		Email            string        `json:"email"`
		ExpiresInSeconds time.Duration `json:"expires_in_seconds,omitempty"`
	}

	decoder := json.NewDecoder(r.Body)
	request := loginRequest{}
	if err := decoder.Decode(&request); err != nil {
		handleError(err, "Failed to decode request", w, http.StatusBadRequest)
		return
	}
	if request.ExpiresInSeconds == 0 {
		request.ExpiresInSeconds = time.Duration(defaultExpiresInHours.Seconds())
	}

	user, err := h.DbQueries.GetUser(r.Context(), request.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(UnauthorizedUserMsg, w, http.StatusUnauthorized)
			return
		}
		handleError(err, "Failed to get user", w, http.StatusInternalServerError)
		return
	}

	authorized, err := auth.CheckPasswordHash(request.Password, user.HashedPassword)
	if err != nil {
		handleError(err, "Failed to check password hash", w, http.StatusInternalServerError)
		return
	}
	if !authorized {
		writeJSON(UnauthorizedUserMsg, w, http.StatusUnauthorized)
		return
	}

	jwt, err := auth.MakeJWT(user.ID, h.JWTSecret, request.ExpiresInSeconds*time.Second)
	if err != nil {
		handleError(err, "Failed to make JWT", w, http.StatusInternalServerError)
		return
	}

	type loginResponse struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Token     string    `json:"token"`
	}
	response := loginResponse{
		Id:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     jwt,
	}
	writeJSON(response, w, http.StatusOK)
}
