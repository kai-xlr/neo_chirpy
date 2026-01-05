package main

import (
	"encoding/json"
	"net/http"

	"github.com/kai-xlr/neo_chirpy/internal/database"
)

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	w.Header().Set("Content-Type", ContentTypeTextPlain)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (a *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var params chirpCreateRequest
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	// Validate chirp body
	if err := ValidateChirpBody(params.Body); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	cleaned := cleanChirp(params.Body)

	chirp, err := a.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleaned,
		UserID: params.UserID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, chirpCreateResponse{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}
