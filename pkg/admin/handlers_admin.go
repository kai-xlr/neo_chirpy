package admin

import (
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/kai-xlr/neo_chirpy/internal/database"
	"github.com/kai-xlr/neo_chirpy/pkg/handlers"
	"github.com/kai-xlr/neo_chirpy/pkg/types"
)

// Config holds configuration needed for admin handlers
type Config struct {
	FileserverHits atomic.Int32
	DB             *database.Queries
	Platform       string
}

// HandlerMetrics handles GET /admin/metrics requests
func (cfg *Config) HandlerMetrics(w http.ResponseWriter, r *http.Request) {
	if !handlers.RequireMethod(w, r, http.MethodGet) {
		return
	}
	w.Header().Set("Content-Type", types.ContentTypeTextHTML)
	fmt.Fprintf(w, `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.FileserverHits.Load())
}

// HandlerReset handles POST /admin/reset requests
func (cfg *Config) HandlerReset(w http.ResponseWriter, r *http.Request) {
	if !handlers.RequireMethod(w, r, http.MethodPost) {
		return
	}
	if cfg.Platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Reset is only allowed in dev environment."))
		return
	}
	cfg.FileserverHits.Store(0)
	err := cfg.DB.Reset(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to reset the database: " + err.Error()))
		return
	}
	w.Header().Set("Content-Type", types.ContentTypeTextPlain)
	w.Write([]byte("Hits reset to 0 and database reset to initial state."))
}
