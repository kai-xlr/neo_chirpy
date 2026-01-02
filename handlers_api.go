package main

import (
	"encoding/json"
	"net/http"
)

type chirpRequest struct {
	Body string `json:"body"`
}

type chirpResponse struct {
	CleanedBody string `json:"cleaned_body"`
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func handlerChirpsValidate(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var params chirpRequest
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	cleaned := cleanChirp(params.Body)

	respondWithJSON(w, http.StatusOK, chirpResponse{
		CleanedBody: cleaned,
	})
}
