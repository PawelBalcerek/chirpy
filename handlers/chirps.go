package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/PawelBalcerek/chirpy/internal/database"
	"github.com/google/uuid"
)

var profaneWords = map[string]struct{}{
	"kerfuffle": {},
	"sharbert":  {},
	"fornax":    {},
}

type CreateChirpHandler struct {
	DbQueries *database.Queries
}

func (h CreateChirpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	type createChirpRequest struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
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
		UserID: request.UserId,
	}
	chirp, err := h.DbQueries.CreateChirp(r.Context(), params)
	if err != nil {
		handleError(err, "Failed to create chirp", w, http.StatusInternalServerError)
	}

	type createChirpResponse struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserId    uuid.UUID `json:"user_id"`
	}
	response := createChirpResponse{
		Id:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	}
	writeJSON(response, w, http.StatusCreated)
}
