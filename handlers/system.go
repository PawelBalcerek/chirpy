package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/PawelBalcerek/chirpy/internal/database"
	"github.com/PawelBalcerek/chirpy/metrics"
)

const metricsPageTemplate = `
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
`

type SystemController struct {
	Metrics   *metrics.Metrics
	DbQueries *database.Queries
	Platform  string
}

func (c *SystemController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, http.StatusText(http.StatusOK))
}

func (c *SystemController) MetricsReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, fmt.Sprintf(metricsPageTemplate, c.Metrics.FileserverHits()))
}

func (c *SystemController) Reset(w http.ResponseWriter, r *http.Request) {
	if c.Platform != "local" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Reset is only allowed in the local environment."))
		return
	}

	if err := c.DbQueries.DeleteUsers(r.Context()); err != nil {
		handleError(err, "Failed to delete users", w, http.StatusInternalServerError)
		return
	}

	c.Metrics.FileserverHitsReset()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Reset succeeded."))
}
