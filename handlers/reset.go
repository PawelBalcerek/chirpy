package handlers

import (
	"net/http"

	"github.com/PawelBalcerek/chirpy/metrics"
)

type ResetHandler struct {
	Metrics   *metrics.Metrics
	DbQueries UserStore
	Platform  string
}

func (h *ResetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.Platform != "local" {
		w.Write([]byte("Reset is only allowed in the local environment."))
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if err := h.DbQueries.DeleteUsers(r.Context()); err != nil {
		handleError(err, "Failed to delete users", w, http.StatusInternalServerError)
		return
	}

	h.Metrics.FileserverHitsReset()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Reset succeed."))
}
