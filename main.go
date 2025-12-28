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

func main() {
	apiCfg := &apiConfig{}

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir(filepathRoot))
	mux.Handle("/", fs)
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fs)))

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

func (a *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (a *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "fileserver_hits %d\n", a.fileserverHits.Load())
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (a *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	a.fileserverHits.Store(0)
	w.Write([]byte("Hits reset to 0\n"))
}
