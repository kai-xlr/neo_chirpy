package main

import (
	"fmt"
	"log"
	"net/http"
)

const (
	port         = 8080
	filepathRoot = "."
)

func main() {
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir(filepathRoot))
	mux.Handle("/", fs)
	mux.Handle("/app/", http.StripPrefix("/app", fs))
	mux.HandleFunc("/healthz", handlerReadiness)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	log.Printf("Serving on port %d", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
