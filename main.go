package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

const (
	port         = 8080
	filepathRoot = "."
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (a *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (a *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits: %d", a.fileserverHits.Load())))
}

func main() {
	apiCfg := &apiConfig{}

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir(filepathRoot))
	mux.Handle("/", fs)

	appHandler := apiCfg.middlewareMetricsInc(
		http.StripPrefix("/app", fs),
	)
	mux.Handle("/app/", appHandler)

	mux.HandleFunc("/healthz", handlerReadiness)
	mux.HandleFunc("/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("/reset", apiCfg.handlerReset)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	log.Printf("Serving on port %d", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
