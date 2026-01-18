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
	"github.com/kai-xlr/neo_chirpy/pkg/admin"
	"github.com/kai-xlr/neo_chirpy/pkg/chirp"
	"github.com/kai-xlr/neo_chirpy/pkg/handlers"
	"github.com/kai-xlr/neo_chirpy/pkg/middleware"
	"github.com/kai-xlr/neo_chirpy/pkg/user"
	"github.com/kai-xlr/neo_chirpy/pkg/webhook"
	_ "github.com/lib/pq"
)

const (
	port         = 8080
	filepathRoot = "."
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	jwtSecret      string
	polkaKey       string

	// Handler configs
	adminConfig      admin.Config
	chirpConfig      chirp.Config
	userConfig       user.Config
	middlewareConfig middleware.Config
	webhookConfig    webhook.Config
}

func main() {
	// Load environment and initialize database
	dbQueries, platform, jwtSecret, polkaKey := initDatabase()

	// Initialize API configuration
	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
		jwtSecret:      jwtSecret,
		polkaKey:       polkaKey,
	}

	// Initialize handler configs
	apiCfg.adminConfig = admin.Config{
		FileserverHits: apiCfg.fileserverHits,
		DB:             dbQueries,
		Platform:       platform,
	}
	apiCfg.chirpConfig = chirp.Config{
		DB:        dbQueries,
		JWTSecret: jwtSecret,
	}
	apiCfg.userConfig = user.Config{
		DB:        dbQueries,
		JWTSecret: jwtSecret,
	}
	apiCfg.middlewareConfig = middleware.Config{
		FileserverHits: apiCfg.fileserverHits,
	}

	// Initialize webhook config
	apiCfg.webhookConfig = webhook.Config{
		DB:       dbQueries,
		PolkaKey: polkaKey,
	}
	apiCfg.chirpConfig = chirp.Config{
		DB:        dbQueries,
		JWTSecret: jwtSecret,
	}
	apiCfg.userConfig = user.Config{
		DB:        dbQueries,
		JWTSecret: jwtSecret,
	}
	apiCfg.middlewareConfig = middleware.Config{
		FileserverHits: apiCfg.fileserverHits,
	}

	// Setup HTTP router
	mux := setupRouter(apiCfg)

	// Start server
	startServer(mux)
}

func initDatabase() (*database.Queries, string, string, string) {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET must be set")
	}

	polkaKey := os.Getenv("POLKA_KEY")
	if polkaKey == "" {
		log.Fatal("POLKA_KEY must be set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}

	return database.New(db), platform, jwtSecret, polkaKey
}

func setupRouter(apiCfg *apiConfig) *http.ServeMux {
	mux := http.NewServeMux()

	// Static file serving
	fs := http.FileServer(http.Dir(filepathRoot))
	mux.Handle("/", fs)
	mux.Handle("/app/", apiCfg.middlewareConfig.MetricsInc(http.StripPrefix("/app", fs)))

	// API endpoints
	mux.HandleFunc("/api/healthz", handlers.HandlerReadiness)
	mux.HandleFunc("/api/chirps", apiCfg.chirpConfig.HandlerCreate)
	mux.HandleFunc("/api/chirps/", apiCfg.chirpConfig.HandlerByID)
	mux.HandleFunc("/api/users", apiCfg.userConfig.HandlerUsers)
	mux.HandleFunc("/api/login", apiCfg.userConfig.HandlerLogin)
	mux.HandleFunc("/api/refresh", apiCfg.userConfig.HandlerRefresh)
	mux.HandleFunc("/api/revoke", apiCfg.userConfig.HandlerRevoke)
	mux.HandleFunc("/api/polka/webhooks", apiCfg.webhookConfig.HandlerPolkaWebhooks)

	// Admin endpoints
	mux.HandleFunc("/admin/metrics", apiCfg.adminConfig.HandlerMetrics)
	mux.HandleFunc("/admin/reset", apiCfg.adminConfig.HandlerReset)

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
