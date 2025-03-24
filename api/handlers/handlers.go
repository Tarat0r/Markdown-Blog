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
func writeJSONError(w http.ResponseWriter, r *http.Request, err error, message string, statusCode int) {
	user_id := r.Context().Value("user_id").(int32)
	log.Println("user:", user_id, "", message, " ", err)
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
func GetIDFromURI(w http.ResponseWriter, r *http.Request, user_id int32) (int32, bool) {
	note_id_int, err := strconv.Atoi(r.PathValue("NoteID"))
	if err != nil {
		writeJSONError(w, r, err, "Invalid note ID", http.StatusBadRequest)
		return 0, false
	}

	note_id := int32(note_id_int)
	log.Println("user:", user_id, " note:", note_id)

	_, err = database.Queries.GetNoteByID(context.Background(), note_id)
	if err != nil {
		writeJSONError(w, r, err, "Note doesn't exist", http.StatusBadRequest)
		return 0, false
	}

	return note_id, true
}

// ComputeSHA256Hash generates a SHA-256 hash of a file
func ComputeSHA256Hash(file multipart.File) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
