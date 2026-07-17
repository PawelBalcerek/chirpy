package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"time"

	"github.com/PawelBalcerek/chirpy/internal/chirp"
	"github.com/PawelBalcerek/chirpy/internal/database"
	"github.com/google/uuid"
)

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

type ChirpController struct {
	DbQueries *database.Queries
}

func (c *ChirpController) Create(w http.ResponseWriter, r *http.Request) {
	userId, ok := GetUserID(r.Context())
	if !ok {
		handleError(nil, "Invalid token", w, http.StatusUnauthorized)
		return
	}

	type createChirpRequest struct {
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(r.Body)
	request := createChirpRequest{}
	if err := decoder.Decode(&request); err != nil {
		handleError(err, "Failed to decode request", w, http.StatusBadRequest)
		return
	}

	chirpBody, err := chirp.NewBody(request.Body)
	if err != nil {
		if errors.Is(err, chirp.ErrBodyTooLong) || errors.Is(err, chirp.ErrBodyEmpty) {
			handleError(nil, err.Error(), w, http.StatusBadRequest)
			return
		}
		handleError(err, "Something went wrong", w, http.StatusInternalServerError)
		return
	}

	params := database.CreateChirpParams{
		Body:   chirpBody.String(),
		UserID: userId,
	}
	dbChirp, err := c.DbQueries.CreateChirp(r.Context(), params)
	if err != nil {
		handleError(err, "Failed to create chirp", w, http.StatusInternalServerError)
		return
	}

	writeJSON(newChirpResponse(dbChirp), w, http.StatusCreated)
}

func (c *ChirpController) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		handleError(err, "Invalid chirp id", w, http.StatusBadRequest)
		return
	}

	chirpVal, err := c.DbQueries.GetChirp(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON("Chirp could not be found.", w, http.StatusNotFound)
			return
		}
		handleError(err, "Failed to get chirp", w, http.StatusInternalServerError)
		return
	}

	writeJSON(newChirpResponse(chirpVal), w, http.StatusOK)
}

func (c *ChirpController) List(w http.ResponseWriter, r *http.Request) {
	rawAuthorId := r.URL.Query().Get("author_id")
	authorId := uuid.Nil
	if rawAuthorId != "" {
		var err error
		authorId, err = uuid.Parse(rawAuthorId)
		if err != nil {
			handleError(err, "Failed to parse author_id", w, http.StatusBadRequest)
			return
		}
	}

	chirps, err := c.DbQueries.GetChirps(r.Context(), authorId)
	if err != nil {
		handleError(err, "Failed to get chirps", w, http.StatusInternalServerError)
		return
	}

	responses := []chirpResponse{}
	for _, chirpVal := range chirps {
		responses = append(responses, newChirpResponse(chirpVal))
	}

	sortQuery := r.URL.Query().Get("sort")
	if sortQuery == "desc" {
		sort.Slice(responses, func(i, j int) bool { return responses[i].CreatedAt.After(responses[j].CreatedAt) })
	}
	writeJSON(responses, w, http.StatusOK)
}

func (c *ChirpController) Delete(w http.ResponseWriter, r *http.Request) {
	userId, ok := GetUserID(r.Context())
	if !ok {
		handleError(nil, "Invalid token", w, http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		handleError(err, "Invalid chirp id", w, http.StatusBadRequest)
		return
	}

	chirpVal, err := c.DbQueries.GetChirp(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON("Chirp could not be found.", w, http.StatusNotFound)
			return
		}
		handleError(err, "Failed to get chirp", w, http.StatusInternalServerError)
		return
	}

	if chirpVal.UserID != userId {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if err = c.DbQueries.DeleteChirp(r.Context(), id); err != nil {
		handleError(err, "Failed to delete chirp", w, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
