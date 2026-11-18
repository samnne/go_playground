package middleware

import (
	"log"
	"net/http"
	"os"
	"protobuffer/trips"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// 1. Pre-handler logic (e.g., logging, auth checks)
		log.Printf("[HTTP] %s %s", r.Method, r.URL.Path)

		// 2. Pass control to the next handler in the chain
		next.ServeHTTP(w, r)

		// 3. Post-handler logic (optional)
	})
}
func RebuildData(next http.Handler) http.Handler {
	var totalQueries int = 0
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entries, err := os.ReadDir("data")
		if err != nil {
			log.Println("Failed to rebuild data")
			next.ServeHTTP(w, r)
		}

		if len(entries) == 0 || len(entries) != 0 && totalQueries >= 100 {
			trips.ZipFiles()
			totalQueries = 0
		}
		totalQueries++
		// 1. Pre-handler logic (e.g., logging, auth checks)
		log.Printf("[HTTP] %s %s", r.Method, r.URL.Path)

		// 2. Pass control to the next handler in the chain
		next.ServeHTTP(w, r)

		// 3. Post-handler logic (optional)
	})
}
