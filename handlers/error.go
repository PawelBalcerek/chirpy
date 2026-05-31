package handlers

import (
	"log"
	"net/http"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func handleAuthorizationError(err error, w http.ResponseWriter) {
	caser := cases.Title(language.English, cases.NoLower)
	handleError(err, caser.String(err.Error()), w, http.StatusUnauthorized)
}

func handleError(err error, errorMsg string, w http.ResponseWriter, statusCode int) {
	if err != nil {
		log.Printf("an error occurred: %v", err)
	}

	type errorResponse struct {
		Error string `json:"error"`
	}

	writeJSON(errorResponse{Error: errorMsg}, w, statusCode)
}
