package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/PawelBalcerek/chirpy/internal/database"
	"github.com/google/uuid"
)

type CreateUserHandler struct {
	DbQueries *database.Queries
}

func (h CreateUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	type createUserRequest struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	request := createUserRequest{}
	if err := decoder.Decode(&request); err != nil {
		handleError(err, "Something went wrong", w, http.StatusInternalServerError)
		return
	}

	user, err := h.DbQueries.CreateUser(r.Context(), request.Email)
	if err != nil {
		handleError(err, "Something went wrong", w, http.StatusInternalServerError)
	}

	type createUserResponse struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"string"`
	}
	response := createUserResponse{
		Id:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	writeJSON(response, w, http.StatusCreated)
}
