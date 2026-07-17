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

const (
	dbURLEnv     = "DB_URL"
	platformEnv  = "PLATFORM"
	jwtSecretEnv = "JWT_SECRET"
	polkaKeyEnv  = "POLKA_KEY"
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
	dbURL := os.Getenv(dbURLEnv)
	if dbURL == "" {
		log.Fatalf("%s must be set", dbURLEnv)
	}
	platform := os.Getenv(platformEnv)
	if platform == "" {
		log.Fatalf("%s must be set", platformEnv)
	}
	jwtSecret := os.Getenv(jwtSecretEnv)
	if jwtSecret == "" {
		log.Fatalf("%s must be set", jwtSecretEnv)
	}
	polkaKey := os.Getenv(polkaKeyEnv)
	if polkaKey == "" {
		log.Fatalf("%s must be set", polkaKeyEnv)
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
	mux.Handle(
		fmt.Sprintf("POST %s/chirps", sCfg.apiPrefix),
		&handlers.CreateChirpHandler{DbQueries: sCfg.dbQueries, JWTSecret: jwtSecret},
	)
	mux.Handle(fmt.Sprintf("GET %s/chirps/{id}", sCfg.apiPrefix), &handlers.GetChirpHandler{DbQueries: sCfg.dbQueries})
	mux.Handle(fmt.Sprintf("GET %s/chirps", sCfg.apiPrefix), &handlers.GetChirpsHandler{DbQueries: sCfg.dbQueries})
	mux.Handle(
		fmt.Sprintf("DELETE %s/chirps/{id}", sCfg.apiPrefix),
		&handlers.DeleteChirpHandler{DbQueries: sCfg.dbQueries, JWTSecret: jwtSecret},
	)
	mux.Handle(fmt.Sprintf("POST %s/users", sCfg.apiPrefix), &handlers.CreateUserHandler{DbQueries: sCfg.dbQueries})
	mux.Handle(
		fmt.Sprintf("PUT %s/users", sCfg.apiPrefix),
		&handlers.UpdateUserHandler{DbQueries: sCfg.dbQueries, JWTSecret: jwtSecret},
	)
	mux.Handle(
		fmt.Sprintf("POST %s/login", sCfg.apiPrefix),
		&handlers.LoginHandler{UserStore: sCfg.dbQueries, TokenStore: sCfg.dbQueries, JWTSecret: jwtSecret},
	)
	mux.Handle(
		fmt.Sprintf("POST %s/refresh", sCfg.apiPrefix),
		&handlers.RefreshHandler{DbQueries: sCfg.dbQueries, JWTSecret: jwtSecret},
	)
	mux.Handle(fmt.Sprintf("POST %s/revoke", sCfg.apiPrefix), &handlers.RevokeHandler{DbQueries: sCfg.dbQueries})
	mux.Handle(
		fmt.Sprintf("POST %s/polka/webhooks", sCfg.apiPrefix),
		&handlers.PolkaWebhookHandler{DbQueries: sCfg.dbQueries, ApiKey: polkaKey},
	)
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
