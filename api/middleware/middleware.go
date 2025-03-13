package middleware

import (
	"log"
	"net/http"
)

// LoggingMiddleware logs incoming requests
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request URI: %s\n", r.RequestURI)
		next(w, r) // Call the next handler
	}
}

// AuthMiddleware authenticates users
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Perform authentication logic here
		isAuthenticated := true // For demonstration purposes

		if isAuthenticated {
			next(w, r) // Call the next handler if authenticated
		} else {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
	}
}
