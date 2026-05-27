package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/PawelBalcerek/chirpy/handlers"
	"github.com/PawelBalcerek/chirpy/internal/database"
	"github.com/PawelBalcerek/chirpy/metrics"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type serverConfig struct {
	appPrefix   string
	appRoot     string
	apiPrefix   string
	adminPrefix string
	port        int
	metrics     *metrics.Metrics
	dbQueries   *database.Queries
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	sCfg := serverConfig{
		appPrefix:   "/app",
		appRoot:     ".",
		apiPrefix:   "/api",
		adminPrefix: "/admin",
		port:        8080,
		metrics:     &metrics.Metrics{},
		dbQueries:   database.New(db),
	}
	mux := http.NewServeMux()
	mux.Handle(
		fmt.Sprintf("%s/", sCfg.appPrefix),
		sCfg.metrics.FileserverHitsInc(http.StripPrefix(sCfg.appPrefix, http.FileServer(http.Dir(sCfg.appRoot)))),
	)
	mux.Handle(fmt.Sprintf("GET %s/healthz", sCfg.apiPrefix), handlers.HealthCheckHandler{})
	mux.Handle(fmt.Sprintf("GET %s/metrics", sCfg.adminPrefix), &handlers.MetricsHandler{Metrics: sCfg.metrics})
	mux.Handle(
		fmt.Sprintf("POST %s/reset", sCfg.adminPrefix),
		&handlers.ResetHandler{Metrics: sCfg.metrics, DbQueries: sCfg.dbQueries, Platform: platform},
	)
	mux.Handle(fmt.Sprintf("POST %s/chirps", sCfg.apiPrefix), &handlers.CreateChirpHandler{DbQueries: sCfg.dbQueries})
	mux.Handle(fmt.Sprintf("GET %s/chirps", sCfg.apiPrefix), &handlers.GetChirpsHandler{DbQueries: sCfg.dbQueries})
	mux.Handle(fmt.Sprintf("POST %s/users", sCfg.apiPrefix), &handlers.CreateUserHandler{DbQueries: sCfg.dbQueries})
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", sCfg.port),
		Handler: mux,
	}
	log.Printf(
		"Serving files from path \"%s\" on http://localhost:%d%s",
		sCfg.appRoot,
		sCfg.port,
		sCfg.appPrefix,
	)
	log.Fatal(s.ListenAndServe())
}
