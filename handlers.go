package main

import "net/http"

// HTTP helper functions

// requireMethod validates the HTTP method and returns false if invalid
func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}
