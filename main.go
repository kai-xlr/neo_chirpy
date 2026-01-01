package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	"github.com/kai-xlr/neo_chirpy/internal/database"
	_ "github.com/lib/pq"
)

const (
	port         = 8080
	filepathRoot = "."
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
}

func main() {
	// Load environment and initialize database
	dbQueries := initDatabase()

	// Initialize API configuration
	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
	}

	// Setup HTTP router
	mux := setupRouter(apiCfg)

	// Start server
	startServer(mux)
}

func initDatabase() *database.Queries {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}

	return database.New(db)
}

func setupRouter(apiCfg *apiConfig) *http.ServeMux {
	mux := http.NewServeMux()

	// Static file serving
	fs := http.FileServer(http.Dir(filepathRoot))
	mux.Handle("/", fs)
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fs)))

	// API endpoints
	mux.HandleFunc("/api/healthz", handlerReadiness)
	mux.HandleFunc("/api/validate_chirp", handlerChirpsValidate)

	// Admin endpoints
	mux.HandleFunc("/admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("/admin/reset", apiCfg.handlerReset)

	return mux
}

func startServer(handler http.Handler) {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
	}

	log.Printf("Serving on port %d", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
