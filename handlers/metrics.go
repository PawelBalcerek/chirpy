package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/PawelBalcerek/chirpy/metrics"
)

type MetricsHandler struct {
	Metrics *metrics.Metrics
}

func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, fmt.Sprintf("Hits: %d", h.Metrics.FileserverHits()))
}
