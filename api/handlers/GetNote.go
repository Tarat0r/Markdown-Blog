package handlers

import (
	"context"
	"net/http"

	"github.com/Tarat0r/Markdown-Blog/database"
)

func GetNote(w http.ResponseWriter, r *http.Request) {
	user_id, ok := r.Context().Value("user_id").(int32)
	if !ok {
		writeJSONError(w, r, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id, ok := GetIDFromURI(w, r, user_id)
	if !ok {
		return
	}

	note, err := database.Queries.GetNoteByID(context.Background(), id)
	if err != nil {
		writeJSONError(w, r, err, "Error getting note", http.StatusInternalServerError)
		return
	}

	ResponseJSON(w, http.StatusOK, note)
}
