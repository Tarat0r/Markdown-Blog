package handlers

import (
	"context"
	"net/http"

	"github.com/Tarat0r/Markdown-Blog/database"
)

type NoteResponse struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	Path      string `json:"path"`
	Content   string `json:"content"`
	Hash      string `json:"hash"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	ContentMd string `json:"content_md,omitempty"` // Omit this field if not needed
}

func GetNote(w http.ResponseWriter, r *http.Request) {
	user_id, ok := r.Context().Value("user_id").(int32)
	if !ok {
		writeJSONError(w, r, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get Authorization header
	mdFlag := r.Header.Get("content_md")

	id, ok := GetIDFromURI(w, r, user_id)
	if !ok {
		return
	}

	note, err := database.Queries.GetNoteByID(context.Background(), id)
	if err != nil {
		writeJSONError(w, r, err, "Error getting note", http.StatusInternalServerError)
		return
	}

	noteResponse := NoteResponse{
		ID: int(note.ID),
		// UserID:    int(note.UserID),
		Path:      string(note.Path),
		Content:   string(note.Content),
		Hash:      string(note.Hash),
		CreatedAt: note.CreatedAt.Time.Format("2006-01-02 15:04:05"),
		UpdatedAt: note.UpdatedAt.Time.Format("2006-01-02 15:04:05"),
	}

	if mdFlag == "true" {
		noteResponse.ContentMd = string(note.ContentMd)
	}

	ResponseJSON(w, http.StatusOK, noteResponse)
}
