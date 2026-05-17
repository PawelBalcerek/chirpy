package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	appPrefix, appRoot, serverPort := "/app", ".", 8080
	mux := http.NewServeMux()
	mux.Handle(fmt.Sprintf("%s/", appPrefix), http.StripPrefix(appPrefix, http.FileServer(http.Dir(appRoot))))
	mux.Handle("/healthz", HealthCheckHandler{})
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", serverPort),
		Handler: mux,
	}
	log.Println(fmt.Sprintf("Serving files from path \"%s\" on http://localhost:%d%s", appRoot, serverPort, appPrefix))
	log.Fatal(s.ListenAndServe())
}

type HealthCheckHandler struct{}

func (h HealthCheckHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, http.StatusText(http.StatusOK))
}
