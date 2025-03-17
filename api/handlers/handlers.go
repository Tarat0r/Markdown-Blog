package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Tarat0r/Markdown-Blog/database"
)

// JSON response structure
type ErrorResponse struct {
	Error string `json:"error"`
}

func ListNotes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user_id, ok := ctx.Value("user_id").(int32)
	log.Println("user_id:", user_id)
	if !ok {
		writeJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	notes, err := database.Queries.ListNotesByUser(context.Background(), user_id)
	if err != nil {
		log.Fatal("Error listing notes:", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notes)
}

func GetNote(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	fmt.Fprintln(w, "Note Id =", id)
}

func CreateNote(w http.ResponseWriter, r *http.Request) {

}

func UpdateNote(w http.ResponseWriter, r *http.Request) {

}

func DeleteNote(w http.ResponseWriter, r *http.Request) {

}

// Helper function to return JSON errors
func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
