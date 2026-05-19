package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

var profaneWords = map[string]struct{}{
	"kerfuffle": {},
	"sharbert":  {},
	"fornax":    {},
}

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
		CleanedBody string `json:"cleaned_body"`
	}

	cleanedBodyElements := []string{}

	for _, e := range strings.Split(request.Body, " ") {
		if _, ok := profaneWords[strings.ToLower(e)]; ok {
			cleanedBodyElements = append(cleanedBodyElements, "****")
			continue
		}
		cleanedBodyElements = append(cleanedBodyElements, e)
	}

	writeJSON(validateChirpResponse{CleanedBody: strings.Join(cleanedBodyElements, " ")}, w, http.StatusOK)
}
