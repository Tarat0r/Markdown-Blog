package handlers

import (
	"context"
	"encoding/json"
	"log"
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
	log.Println("UserID:", user_id, ", NoteID:", note_id)

	_, err = database.Queries.GetNoteByID(context.Background(), note_id)
	if err != nil {
		writeJSONError(w, r, err, "Note doesn't exist", http.StatusBadRequest)
		return 0, false
	}

	return note_id, true
}
