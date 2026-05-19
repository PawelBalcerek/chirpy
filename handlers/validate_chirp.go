package handlers

import (
	"encoding/json"
	"net/http"
)

type ValidateChirpHandler struct{}

func (h ValidateChirpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	type validateChirpRequest struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	request := validateChirpRequest{}
	if err := decoder.Decode(&request); err != nil {
		handleError(err, "Something went wrong", w, http.StatusInternalServerError)
		return
	}

	if len(request.Body) > 140 {
		handleError(nil, "Chirp is too long", w, http.StatusBadRequest)
		return
	}

	type validateChirpResponse struct {
		Valid bool `json:"valid"`
	}

	writeJSON(validateChirpResponse{Valid: true}, w, http.StatusOK)
}
