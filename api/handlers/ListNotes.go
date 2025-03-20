package handlers

import (
	"context"
	"net/http"

	"github.com/Tarat0r/Markdown-Blog/database"
	db "github.com/Tarat0r/Markdown-Blog/database/sqlc"
)

func ListNotes(w http.ResponseWriter, r *http.Request) {
	user_id, ok := r.Context().Value("user_id").(int32)
	if !ok {
		writeJSONError(w, r, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	notes, err := database.Queries.ListNotesByUser(context.Background(), user_id)
	if err != nil {
		writeJSONError(w, r, err, "Error listing notes", http.StatusInternalServerError)
		return
	} else if notes == nil {
		notes = make([]db.ListNotesByUserRow, 0)
	}
	ResponseJSON(w, http.StatusOK, notes)
}
