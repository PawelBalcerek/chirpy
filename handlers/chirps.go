package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/PawelBalcerek/chirpy/internal/auth"
	"github.com/PawelBalcerek/chirpy/internal/database"
	"github.com/google/uuid"
)

var profaneWords = map[string]struct{}{
	"kerfuffle": {},
	"sharbert":  {},
	"fornax":    {},
}

type chirpResponse struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

func newChirpResponse(chirp database.Chirp) chirpResponse {
	return chirpResponse{
		Id:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	}
}

type CreateChirpHandler struct {
	DbQueries *database.Queries
	JWTSecret string
}

func (h CreateChirpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		handleAuthorizationError(err, w)
		return
	}

	userId, err := auth.ValidateJWT(token, h.JWTSecret)
	if err != nil {
		handleError(err, "Invalid token", w, http.StatusUnauthorized)
		return
	}

	type createChirpRequest struct {
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(r.Body)
	request := createChirpRequest{}
	if err := decoder.Decode(&request); err != nil {
		handleError(err, "Something went wrong", w, http.StatusInternalServerError)
		return
	}

	if len(request.Body) > 140 {
		handleError(nil, "Chirp is too long", w, http.StatusBadRequest)
		return
	}

	cleanedBodyElements := []string{}
	for e := range strings.SplitSeq(request.Body, " ") {
		if _, ok := profaneWords[strings.ToLower(e)]; ok {
			cleanedBodyElements = append(cleanedBodyElements, "****")
			continue
		}
		cleanedBodyElements = append(cleanedBodyElements, e)
	}

	params := database.CreateChirpParams{
		Body:   request.Body,
		UserID: userId,
	}
	chirp, err := h.DbQueries.CreateChirp(r.Context(), params)
	if err != nil {
		handleError(err, "Failed to create chirp", w, http.StatusInternalServerError)
		return
	}

	writeJSON(newChirpResponse(chirp), w, http.StatusCreated)
}

type GetChirpHandler struct {
	DbQueries *database.Queries
}

func (h GetChirpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		handleError(err, "Invalid chirp id", w, http.StatusBadRequest)
		return
	}

	chirp, err := h.DbQueries.GetChirp(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON("Chirp could not be found.", w, http.StatusNotFound)
			return
		}
		handleError(err, "Failed to get chirp", w, http.StatusInternalServerError)
		return
	}

	writeJSON(newChirpResponse(chirp), w, http.StatusOK)
}

type GetChirpsHandler struct {
	DbQueries *database.Queries
}

func (h GetChirpsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	chirps, err := h.DbQueries.GetChirps(r.Context())
	if err != nil {
		handleError(err, "Failed to get chirps", w, http.StatusInternalServerError)
		return
	}

	responses := []chirpResponse{}
	for _, chirp := range chirps {
		responses = append(responses, newChirpResponse(chirp))
	}
	writeJSON(responses, w, http.StatusOK)
}

type DeleteChirpHandler struct {
	DbQueries *database.Queries
	JWTSecret string
}

func (h DeleteChirpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		handleAuthorizationError(err, w)
		return
	}

	userId, err := auth.ValidateJWT(token, h.JWTSecret)
	if err != nil {
		handleError(err, "Invalid token", w, http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		handleError(err, "Invalid chirp id", w, http.StatusBadRequest)
		return
	}

	chirp, err := h.DbQueries.GetChirp(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON("Chirp could not be found.", w, http.StatusNotFound)
			return
		}
		handleError(err, "Failed to get chirp", w, http.StatusInternalServerError)
		return
	}

	if chirp.UserID != userId {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if err = h.DbQueries.DeleteChirp(r.Context(), id); err != nil {
		handleError(err, "Failed to delete chirp", w, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
