package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"ticket-system/internal/handlers"
	"ticket-system/internal/middleware"
	"ticket-system/internal/store"
)

// healthHandler responds to GET /health.
// This must stay publicly reachable with no auth once deployed.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "ticket-system.db"
	}

	s, err := store.New(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	h := handlers.New(s)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("POST /auth/register", h.Register)
	mux.HandleFunc("POST /auth/login", h.Login)

	mux.HandleFunc("POST /tickets", middleware.RequireAuth(h.CreateTicket))
	mux.HandleFunc("GET /tickets", middleware.RequireAuth(h.ListTickets))
	mux.HandleFunc("GET /tickets/{id}", middleware.RequireAuth(h.GetTicket))
	mux.HandleFunc("PATCH /tickets/{id}/status", middleware.RequireAuth(h.UpdateTicketStatus))

	handler := middleware.CORS(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}
