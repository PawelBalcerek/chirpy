package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/PawelBalcerek/chirpy/internal/auth"
	"github.com/PawelBalcerek/chirpy/internal/database"
)

const (
	UnauthorizedUserMsg = "Incorrect email or password"
)

type LoginHandler struct {
	DbQueries *database.Queries
}

func (h LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	type loginRequest struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	request := loginRequest{}
	if err := decoder.Decode(&request); err != nil {
		handleError(err, "Failed to decode request", w, http.StatusBadRequest)
		return
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

	response := UserResponse{
		Id:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	writeJSON(response, w, http.StatusOK)
}
