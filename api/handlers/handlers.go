package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/Tarat0r/Markdown-Blog/database"
)

// JSON response structure
type ErrorResponse struct {
	Error string `json:"error"`
}

// Helper function to return JSON errors
// writeJSONError sends a JSON error response to the client.
// Parameters:
// - w: The HTTP response writer used to send the response.
// - r: The HTTP request object containing context and other details.
// - err: The error object to log for debugging purposes.
// - message: A user-friendly error message to include in the response.
// - statusCode: The HTTP status code to set for the response.
func writeJSONError(w http.ResponseWriter, r *http.Request, err error, message string, statusCode int) {
	userIDValue := r.Context().Value("contextUserID")
	contextUserID, ok := userIDValue.(int32)
	if !ok {
		log.Println("Error: contextUserIDnot found or invalid type in context")
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Internal server error"})
		return
	}
	log.Printf("user: %d, message: %s, error: %v", contextUserID, message, err)
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

// Helper function to return JSON response
func ResponseJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

// Helper function to get NoteID from the uri
func GetIDFromURI(w http.ResponseWriter, r *http.Request, contextUserID int32) (int32, bool) {
	noteIDint, err := strconv.Atoi(r.PathValue("NoteID"))
	if err != nil {
		writeJSONError(w, r, err, "Invalid note ID", http.StatusBadRequest)
		return 0, false
	}

	noteID := int32(noteIDint)
	log.Println("user:", contextUserID, " note:", noteID)

	_, err = database.Queries.GetNoteByID(context.Background(), noteID)
	if err != nil {
		writeJSONError(w, r, err, "Note doesn't exist", http.StatusBadRequest)
		return 0, false
	}

	return noteID, true
}

// ComputeSHA256Hash generates a SHA-256 hash of a file.
// Parameters:
// - file: The multipart.File object representing the file to hash.
// Returns:
// - A string containing the hexadecimal representation of the SHA-256 hash.
// - An error if there is an issue reading the file or computing the hash.
func ComputeSHA256Hash(file multipart.File) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
