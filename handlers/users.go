package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/PawelBalcerek/chirpy/internal/auth"
	"github.com/PawelBalcerek/chirpy/internal/database"
	"github.com/google/uuid"
)

type UserResponse struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type CreateUserHandler struct {
	DbQueries *database.Queries
}

func (h CreateUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	type createUserRequest struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	request := createUserRequest{}
	if err := decoder.Decode(&request); err != nil {
		handleError(err, "Failed to decode request", w, http.StatusBadRequest)
		return
	}

	hashedPassword, err := auth.HashPassword(request.Password)
	if err != nil {
		handleError(err, "Failed to hash password", w, http.StatusInternalServerError)
		return
	}

	params := database.CreateUserParams{
		Email:          request.Email,
		HashedPassword: hashedPassword,
	}
	user, err := h.DbQueries.CreateUser(r.Context(), params)
	if err != nil {
		handleError(err, "Failed to create user", w, http.StatusInternalServerError)
		return
	}

	response := UserResponse{
		Id:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	writeJSON(response, w, http.StatusCreated)
}
