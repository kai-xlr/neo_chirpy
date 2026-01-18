package middleware

import (
	"net/http"
	"sync/atomic"
)

// Config holds configuration needed for middleware
type Config struct {
	FileserverHits atomic.Int32
}

// MetricsInc increments the file server hits counter
func (cfg *Config) MetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
