package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"ticket-system/internal/auth"
)

// contextKey is a private type so our context key can never collide with
// keys set by other packages.
type contextKey string

const userIDKey contextKey = "userID"

// RequireAuth wraps a handler so it only runs if the request has a valid
// "Authorization: Bearer <token>" header. On success, the authenticated
// user's ID is attached to the request context for the handler to read.
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			writeUnauthorized(w)
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeUnauthorized(w)
			return
		}

		userID, err := auth.ParseToken(parts[1])
		if err != nil {
			writeUnauthorized(w)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}

// GetUserID reads the authenticated user's ID out of the request context.
// Handlers call this after RequireAuth has already validated the token.
func GetUserID(r *http.Request) (int64, bool) {
	id, ok := r.Context().Value(userIDKey).(int64)
	return id, ok
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{"error": "missing or invalid authorization token"})
}
