package middleware

import "net/http"

// CORS allows a browser-based frontend (running on a different origin,
// e.g. opened as a local file or served on another port) to call this API.
// It's deliberately permissive (Allow-Origin: *) since this is a small
// assignment project, not a service handling sensitive third-party data.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Browsers send an OPTIONS "preflight" request before the real one
		// for any request with custom headers (like our Authorization header).
		// We answer it here, before it ever reaches the router, since our
		// routes are registered per-method and would otherwise 404 on OPTIONS.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
