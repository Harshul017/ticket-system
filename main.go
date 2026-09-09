package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// healthHandler responds to GET /health.
// This is the endpoint the assignment says must be publicly reachable
// once deployed, with no authentication required.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	mux := http.NewServeMux()

	// Go 1.22+ lets you specify the HTTP method directly in the pattern.
	mux.HandleFunc("GET /health", healthHandler)

	log.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
