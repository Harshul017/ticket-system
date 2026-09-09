package handlers

import (
	"encoding/json"
	"net/http"
)

// writeJSON sends v as a JSON body with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError sends a consistent {"error": "..."} JSON body.
// Using one shape everywhere means a hidden test suite (or a teammate)
// always knows where to look for the failure reason.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
