package handlers

import (
	"net/http"

	"github.com/kai-xlr/neo_chirpy/pkg/types"
)

// HandlerReadiness responds to GET /api/healthz with a simple "OK" message
func HandlerReadiness(w http.ResponseWriter, r *http.Request) {
	if !RequireMethod(w, r, http.MethodGet) {
		return
	}
	w.Header().Set("Content-Type", types.ContentTypeTextPlain)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
