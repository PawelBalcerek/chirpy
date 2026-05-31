package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/PawelBalcerek/chirpy/internal/auth"
	"github.com/PawelBalcerek/chirpy/internal/database"
	"github.com/google/uuid"
)

type userRequest struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type userResponse struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func newUserResponse(user database.User) userResponse {
	return userResponse{
		Id:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
}

type CreateUserHandler struct {
	DbQueries *database.Queries
}

func (h CreateUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	request := userRequest{}
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

	writeJSON(newUserResponse(user), w, http.StatusCreated)
}

type UpdateUserHandler struct {
	DbQueries *database.Queries
	JWTSecret string
}

func (h UpdateUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		handleBearerTokenError(err, w)
		return
	}
	userId, err := auth.ValidateJWT(token, h.JWTSecret)
	if err != nil {
		handleError(err, "Invalid token", w, http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	request := userRequest{}
	if err := decoder.Decode(&request); err != nil {
		handleError(err, "Failed to decode request", w, http.StatusBadRequest)
		return
	}

	hashedPassword, err := auth.HashPassword(request.Password)
	if err != nil {
		handleError(err, "Failed to hash password", w, http.StatusInternalServerError)
		return
	}

	params := database.UpdateUserParams{
		Email:          request.Email,
		HashedPassword: hashedPassword,
		ID:             userId,
	}
	user, err := h.DbQueries.UpdateUser(r.Context(), params)
	if err != nil {
		handleError(err, "Failed to create user", w, http.StatusInternalServerError)
		return
	}

	writeJSON(newUserResponse(user), w, http.StatusOK)
}
