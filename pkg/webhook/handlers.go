package webhook

import (
	"encoding/json"
	"net/http"

	"github.com/kai-xlr/neo_chirpy/internal/auth"
	"github.com/kai-xlr/neo_chirpy/internal/database"
	"github.com/kai-xlr/neo_chirpy/pkg/handlers"
	"github.com/kai-xlr/neo_chirpy/pkg/types"
)

// Config holds configuration needed for webhook handlers
type Config struct {
	DB       *database.Queries
	PolkaKey string
}

// HandlerPolkaWebhooks handles POST /api/polka/webhooks requests
func (cfg *Config) HandlerPolkaWebhooks(w http.ResponseWriter, r *http.Request) {
	if !handlers.RequireMethod(w, r, http.MethodPost) {
		return
	}

	// Validate API key
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		handlers.RespondWithError(w, http.StatusUnauthorized, auth.ErrUnauthorized.Error(), err)
		return
	}

	if apiKey != cfg.PolkaKey {
		handlers.RespondWithError(w, http.StatusUnauthorized, auth.ErrUnauthorized.Error(), auth.ErrUnauthorized)
		return
	}

	// Parse JSON from request body
	var request types.WebhookRequest
	decodeErr := json.NewDecoder(r.Body).Decode(&request)
	if decodeErr != nil {
		handlers.RespondWithError(w, http.StatusInternalServerError, types.ErrMsgDecodeParams, decodeErr)
		return
	}

	// If event is not user.upgraded, respond with 204 (we don't care about other events)
	if request.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Upgrade user to Chirpy Red
	_, err = cfg.DB.UpgradeUserToChirpyRed(r.Context(), request.Data.UserID)
	if err != nil {
		if err.Error() == "no rows in result set" || err.Error() == "sql: no rows in result set" {
			handlers.RespondWithError(w, http.StatusNotFound, "User not found", err)
		} else {
			handlers.RespondWithError(w, http.StatusInternalServerError, "Couldn't upgrade user", err)
		}
		return
	}

	// Return 204 No Content for successful upgrade
	w.WriteHeader(http.StatusNoContent)
}
