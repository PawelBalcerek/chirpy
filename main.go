package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/PawelBalcerek/chirpy/handlers"
	"github.com/PawelBalcerek/chirpy/internal/auth"
	"github.com/PawelBalcerek/chirpy/internal/database"
	"github.com/PawelBalcerek/chirpy/metrics"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	dbURLEnv     = "DB_URL"
	platformEnv  = "PLATFORM"
	tokenSecretEnv = "JWT_SECRET"
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
	tokenSecret := os.Getenv(tokenSecretEnv)
	if tokenSecret == "" {
		log.Fatalf("%s must be set", tokenSecretEnv)
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

	chirpCtrl := &handlers.ChirpController{ChirpStore: sCfg.dbQueries}
	userCtrl := &handlers.UserController{UserStore: sCfg.dbQueries, TokenStore: sCfg.dbQueries, TokenSecret: auth.TokenSecret(tokenSecret)}
	polkaCtrl := &handlers.PolkaController{UserStore: sCfg.dbQueries}
	systemCtrl := &handlers.SystemController{
		Metrics:   sCfg.metrics,
		UserStore: sCfg.dbQueries,
		Platform:  handlers.Platform(platform),
	}

	mux.Handle(
		fmt.Sprintf("%s/", sCfg.appPrefix),
		sCfg.metrics.FileserverHitsInc(http.StripPrefix(sCfg.appPrefix, http.FileServer(http.Dir(sCfg.appRoot)))),
	)
	mux.HandleFunc(fmt.Sprintf("GET %s/healthz", sCfg.apiPrefix), systemCtrl.HealthCheck)
	mux.HandleFunc(fmt.Sprintf("GET %s/metrics", sCfg.adminPrefix), systemCtrl.MetricsReport)
	mux.HandleFunc(fmt.Sprintf("POST %s/reset", sCfg.adminPrefix), systemCtrl.Reset)

	mux.Handle(
		fmt.Sprintf("POST %s/chirps", sCfg.apiPrefix),
		handlers.RequireJWT(auth.TokenSecret(tokenSecret))(http.HandlerFunc(chirpCtrl.Create)),
	)
	mux.HandleFunc(fmt.Sprintf("GET %s/chirps/{id}", sCfg.apiPrefix), chirpCtrl.Get)
	mux.HandleFunc(fmt.Sprintf("GET %s/chirps", sCfg.apiPrefix), chirpCtrl.List)
	mux.Handle(
		fmt.Sprintf("DELETE %s/chirps/{id}", sCfg.apiPrefix),
		handlers.RequireJWT(auth.TokenSecret(tokenSecret))(http.HandlerFunc(chirpCtrl.Delete)),
	)

	mux.HandleFunc(fmt.Sprintf("POST %s/users", sCfg.apiPrefix), userCtrl.Create)
	mux.Handle(
		fmt.Sprintf("PUT %s/users", sCfg.apiPrefix),
		handlers.RequireJWT(auth.TokenSecret(tokenSecret))(http.HandlerFunc(userCtrl.Update)),
	)
	mux.HandleFunc(fmt.Sprintf("POST %s/login", sCfg.apiPrefix), userCtrl.Login)
	mux.HandleFunc(fmt.Sprintf("POST %s/refresh", sCfg.apiPrefix), userCtrl.Refresh)
	mux.HandleFunc(fmt.Sprintf("POST %s/revoke", sCfg.apiPrefix), userCtrl.Revoke)

	mux.Handle(
		fmt.Sprintf("POST %s/polka/webhooks", sCfg.apiPrefix),
		handlers.RequireApiKey(polkaKey)(http.HandlerFunc(polkaCtrl.ReceiveWebhook)),
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
