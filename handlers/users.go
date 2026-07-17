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
	UnauthorizedUserMsg = "Incorrect email or password"

	jwtExpiresIn          = 1 * time.Hour
	refreshTokenExpiresIn = 60 * 24 * time.Hour
)

type userRequest struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type userResponse struct {
	Id          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

func newUserResponse(user database.User) userResponse {
	return userResponse{
		Id:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}
}

type UserController struct {
	DbQueries *database.Queries
	JWTSecret string
}

func (c *UserController) Create(w http.ResponseWriter, r *http.Request) {
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
	user, err := c.DbQueries.CreateUser(r.Context(), params)
	if err != nil {
		handleError(err, "Failed to create user", w, http.StatusInternalServerError)
		return
	}

	writeJSON(newUserResponse(user), w, http.StatusCreated)
}

func (c *UserController) Update(w http.ResponseWriter, r *http.Request) {
	userId, ok := GetUserID(r.Context())
	if !ok {
		handleError(nil, "Invalid token", w, http.StatusUnauthorized)
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
	user, err := c.DbQueries.UpdateUser(r.Context(), params)
	if err != nil {
		handleError(err, "Failed to update user", w, http.StatusInternalServerError)
		return
	}

	writeJSON(newUserResponse(user), w, http.StatusOK)
}

func (c *UserController) Login(w http.ResponseWriter, r *http.Request) {
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

	user, err := c.DbQueries.GetUser(r.Context(), request.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			handleError(nil, UnauthorizedUserMsg, w, http.StatusUnauthorized)
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
		handleError(nil, UnauthorizedUserMsg, w, http.StatusUnauthorized)
		return
	}

	jwt, err := auth.MakeJWT(user.ID, c.JWTSecret, jwtExpiresIn)
	if err != nil {
		handleError(err, "Failed to make JWT", w, http.StatusInternalServerError)
		return
	}

	refreshToken := auth.MakeRefreshToken()

	params := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(refreshTokenExpiresIn),
	}
	if _, err := c.DbQueries.CreateRefreshToken(r.Context(), params); err != nil {
		handleError(err, "Failed to create refresh token", w, http.StatusInternalServerError)
		return
	}

	type loginResponse struct {
		Id           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		IsChirpyRed  bool      `json:"is_chirpy_red"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}
	response := loginResponse{
		Id:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		IsChirpyRed:  user.IsChirpyRed,
		Token:        jwt,
		RefreshToken: refreshToken,
	}
	writeJSON(response, w, http.StatusOK)
}

func (c *UserController) Refresh(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		handleAuthorizationError(err, w)
		return
	}

	refreshToken, err := c.DbQueries.GetRefreshToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			handleError(nil, "Unknown refresh token", w, http.StatusUnauthorized)
			return
		}
		handleError(err, "Failed to obtain refresh token", w, http.StatusInternalServerError)
		return
	}

	if refreshToken.RevokedAt.Valid {
		handleError(nil, "Refresh token has been revoked", w, http.StatusUnauthorized)
		return
	}

	if !refreshToken.ExpiresAt.After(time.Now()) {
		handleError(nil, "Expired refresh token", w, http.StatusUnauthorized)
		return
	}

	jwt, err := auth.MakeJWT(refreshToken.UserID, c.JWTSecret, jwtExpiresIn)
	if err != nil {
		handleError(err, "Failed to make JWT", w, http.StatusInternalServerError)
		return
	}

	type refreshResponse struct {
		Token string `json:"token"`
	}
	writeJSON(refreshResponse{Token: jwt}, w, http.StatusOK)
}

func (c *UserController) Revoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		handleAuthorizationError(err, w)
		return
	}

	if err := c.DbQueries.RevokeRefreshToken(r.Context(), token); err != nil {
		handleError(err, "Failed to revoke refresh token", w, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
