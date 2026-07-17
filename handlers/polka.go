package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/PawelBalcerek/chirpy/internal/database"
	"github.com/google/uuid"
)

type PolkaController struct {
	DbQueries *database.Queries
}

func (c *PolkaController) ReceiveWebhook(w http.ResponseWriter, r *http.Request) {

	type polkaWebhookRequest struct {
		Event string `json:"event"`
		Data  struct {
			UserId uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	request := polkaWebhookRequest{}
	if err := decoder.Decode(&request); err != nil {
		handleError(err, "Invalid request body", w, http.StatusBadRequest)
		return
	}

	if request.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if _, err := c.DbQueries.MakeUserChirpyRed(r.Context(), request.Data.UserId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			handleError(nil, "User could not be found", w, http.StatusNotFound)
			return
		}
		handleError(err, "Failed to make user chirpy red", w, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
