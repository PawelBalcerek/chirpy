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
	apiPrefix string
	port      int
	metrics   *metrics.Metrics
}

func main() {
	sCfg := serverConfig{
		appPrefix: "/app",
		appRoot:   ".",
		apiPrefix: "/api",
		port:      8080,
		metrics:   &metrics.Metrics{},
	}
	mux := http.NewServeMux()
	mux.Handle(
		fmt.Sprintf("%s/", sCfg.appPrefix),
		sCfg.metrics.FileserverHitsInc(http.StripPrefix(sCfg.appPrefix, http.FileServer(http.Dir(sCfg.appRoot)))),
	)
	mux.Handle(fmt.Sprintf("GET %s/healthz", sCfg.apiPrefix), handlers.HealthCheckHandler{})
	mux.Handle(fmt.Sprintf("GET %s/metrics", sCfg.apiPrefix), &handlers.MetricsHandler{Metrics: sCfg.metrics})
	mux.Handle(fmt.Sprintf("POST %s/reset", sCfg.apiPrefix), &handlers.ResetHandler{Metrics: sCfg.metrics})
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
