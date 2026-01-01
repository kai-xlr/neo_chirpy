package main

import "strings"

func cleanChirp(body string) string {
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
