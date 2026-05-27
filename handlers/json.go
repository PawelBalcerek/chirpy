package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

func writeJSON(payload any, w http.ResponseWriter, statusCode int) {
	responseBody, err := json.Marshal(payload)
	if err != nil {
		log.Printf("an error occurred during json.Marshal: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(responseBody)
}
