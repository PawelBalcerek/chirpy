package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PawelBalcerek/chirpy/handlers"
	"github.com/PawelBalcerek/chirpy/metrics"
)

type serverConfig struct {
	appPrefix string
	appRoot   string
	port      int
	metrics   *metrics.Metrics
}

func main() {
	sCfg := serverConfig{
		appPrefix: "/app",
		appRoot:   ".",
		port:      8080,
		metrics:   &metrics.Metrics{},
	}
	mux := http.NewServeMux()
	mux.Handle(
		fmt.Sprintf("%s/", sCfg.appPrefix),
		sCfg.metrics.FileserverHitsInc(http.StripPrefix(sCfg.appPrefix, http.FileServer(http.Dir(sCfg.appRoot)))),
	)
	mux.Handle("/healthz", handlers.HealthCheckHandler{})
	mux.Handle("/metrics", &handlers.MetricsHandler{Metrics: sCfg.metrics})
	mux.Handle("/reset", &handlers.ResetHandler{Metrics: sCfg.metrics})
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", sCfg.port),
		Handler: mux,
	}
	log.Println(fmt.Sprintf(
		"Serving files from path \"%s\" on http://localhost:%d%s",
		sCfg.appRoot,
		sCfg.port,
		sCfg.appPrefix,
	))
	log.Fatal(s.ListenAndServe())
}
