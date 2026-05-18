package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/PawelBalcerek/chirpy/metrics"
)

const MetricsPageTemplate = `
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
`

type MetricsHandler struct {
	Metrics *metrics.Metrics
}

func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, fmt.Sprintf(MetricsPageTemplate, h.Metrics.FileserverHits()))
}
