package middleware

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"regexp"

	"github.com/Tarat0r/Markdown-Blog/database"
)

// Context key type to avoid conflicts
type contextKey string

const UserIDKey contextKey = "contextUserID"

// JSON response structure
type ErrorResponse struct {
	Error string `json:"error"`
}

// LoggingMiddleware logs incoming requests
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// log.Printf("user: %d  Request: %s %s\n", r.Context().Value("contextUserID").(int32), r.Method, r.RequestURI)
		next(w, r) // Call the next handler
	}
}

// AuthMiddleware authenticates users
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Get Authorization header
		token := r.Header.Get("Authorization")
		if token == "" {
			// writeJSONError(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Check token format
		if !isValidToken(token) {
			// writeJSONError(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		// Check token in database
		id, err := database.Queries.GetIDByToken(context.Background(), token)

		if errors.Is(err, sql.ErrNoRows) {
			// writeJSONError(w, "Invalid API token", http.StatusUnauthorized)
			return
		}

		if err != nil {
			// log.Println("Database error:", err)
			// writeJSONError(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, "contextUserID", id)

		next(w, r.WithContext(ctx)) // Call the next handler if authenticated
	}
}

// Helper function to return JSON errors
// func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
// 	w.WriteHeader(statusCode)
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
// }

// Helper function to valid API Token
func isValidToken(token string) bool {
	match, _ := regexp.MatchString(`^[a-zA-Z0-9]{64}$`, token)
	return match
}
