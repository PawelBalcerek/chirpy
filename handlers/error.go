package handlers

import (
	"log"
	"net/http"
)

func handleError(err error, errorMsg string, w http.ResponseWriter, statusCode int) {
	if err != nil {
		log.Printf("an error occurred: %v", err)
	}

	type errorResponse struct {
		Error string `json:"error"`
	}

	writeJSON(errorResponse{Error: errorMsg}, w, statusCode)
}
