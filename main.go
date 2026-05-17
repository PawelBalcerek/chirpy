package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	root, port := ".", 8080
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(root)))
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	log.Println(fmt.Sprintf("Serving files from path \"%s\" on localhost:%d", root, port))
	log.Fatal(s.ListenAndServe())
}
