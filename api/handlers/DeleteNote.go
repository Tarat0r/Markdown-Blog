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
	contextUserID, ok := r.Context().Value("contextUserID").(int32)
	if !ok {
		// writeJSONError(w, r, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	noteID, ok := GetIDFromURI(w, r, contextUserID)
	if !ok {
		return
	}

	params := db.DeleteNoteParams{
		UserID: contextUserID,
		ID:     noteID,
	}

	path, err := database.Queries.DeleteNote(context.Background(), params)
	if err != nil {
		// writeJSONError(w, r, err, "Error deleting note", http.StatusInternalServerError)
		return
	}
	pathJSON := map[string]string{"message": "Note deleted successfully", "path": path}
	log.Println("path:", pathJSON)
	ResponseJSON(w, http.StatusOK, pathJSON)
}
