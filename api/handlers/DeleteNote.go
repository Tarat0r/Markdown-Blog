package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/Tarat0r/Markdown-Blog/database"
	db "github.com/Tarat0r/Markdown-Blog/database/sqlc"
)

// TODO Delete images if not used
func DeleteNote(w http.ResponseWriter, r *http.Request) {
	user_id, ok := r.Context().Value("user_id").(int32)
	if !ok {
		writeJSONError(w, r, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	note_id, ok := GetIDFromURI(w, r, user_id)
	if !ok {
		return
	}

	params := db.DeleteNoteParams{
		UserID: user_id,
		ID:     note_id,
	}

	path, err := database.Queries.DeleteNote(context.Background(), params)
	if err != nil {
		writeJSONError(w, r, err, "Error deleting note", http.StatusInternalServerError)
		return
	}
	pathJSON := map[string]string{"message": "Note deleted successfully", "path": path}
	log.Println("path:", pathJSON)
	ResponseJSON(w, http.StatusOK, pathJSON)
}
