package chirp

import "strings"

// CleanChirp removes profanity from chirp text
func CleanChirp(body string) string {
	banned := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	words := strings.Fields(body)

	for i, word := range words {
		lower := strings.ToLower(word)
		if _, found := banned[lower]; found {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}
