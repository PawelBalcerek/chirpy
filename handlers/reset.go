package handlers

import (
	"net/http"

	"github.com/PawelBalcerek/chirpy/metrics"
)

type ResetHandler struct {
	Metrics *metrics.Metrics
}

func (h *ResetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Metrics.FileserverHitsReset()
	w.WriteHeader(http.StatusOK)
}
