package metrics

import (
	"net/http"
	"sync/atomic"
)

type Metrics struct {
	fileserverHits atomic.Int32
}

func (m *Metrics) FileserverHits() int32 {
	return m.fileserverHits.Load()
}

func (m *Metrics) FileserverHitsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (m *Metrics) FileserverHitsReset() {
	m.fileserverHits.Store(0)
}
