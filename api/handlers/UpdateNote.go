package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/Tarat0r/Markdown-Blog/database"
	db "github.com/Tarat0r/Markdown-Blog/database/sqlc"
)

func UpdateNote(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	user_id, ok := r.Context().Value("user_id").(int32)
	if !ok {
		writeJSONError(w, r, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse multipart form (Max 50MB)
	err := r.ParseMultipartForm(50 << 20)
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	// Extract JSON metadata
	var req UploadRequest
	metadata := r.FormValue("metadata")
	if err := json.Unmarshal([]byte(metadata), &req); err != nil {
		writeJSONError(w, r, err, "Invalid JSON metadata", http.StatusBadRequest)
		return
	}

	// Handle Image Uploads
	images, err := ImageUploadHandler(w, r, req, user_id)
	if err != nil {
		log.Println("user:", user_id, "", err)
		writeJSONError(w, r, err, "Failed to save images", http.StatusInternalServerError)
		return
	}

	// Handle Markdown file upload
	markdownFiles := r.MultipartForm.File["markdown"]
	if len(markdownFiles) != 1 {
		writeJSONError(w, r, nil, "Exactly one Markdown file is required", http.StatusBadRequest)
		return
	}

	mdFile, header, err := r.FormFile("markdown")
	if err != nil {
		writeJSONError(w, r, err, "Markdown file is required", http.StatusBadRequest)
		return
	}
	defer mdFile.Close()

	// Check MIME type of Markdown file
	buffer := make([]byte, 512)
	if _, err := mdFile.Read(buffer); err != nil {
		log.Println("user:", user_id, "", "Failed to read file", " ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, err := mdFile.Seek(0, io.SeekStart); err != nil {
		log.Println("user:", user_id, "", "Failed to reset file pointer", " ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	mimeType := http.DetectContentType(buffer)
	if !strings.HasPrefix(mimeType, "text/plain") {
		writeJSONError(w, r, nil, "Invalid markdown file type", http.StatusBadRequest)
		return
	}

	var note_params db.UpdateNoteParams

	note_params.Path = req.Path
	note_params.UserID = user_id
	note_params.ID, ok = GetIDFromURI(w, r, user_id)
	if !ok {
		return
	}
	// Compute SHA-256 Hash of Markdown file
	note_params.Hash, err = ComputeSHA256Hash(mdFile)
	if err != nil {
		log.Println("user:", user_id, "", "Failed to compute markdown file hash", " ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Reset file pointer to the beginning before reading the content
	if _, err := mdFile.Seek(0, io.SeekStart); err != nil {
		log.Println("user:", user_id, "", "Failed to reset file pointer", " ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	mdContent, err := io.ReadAll(mdFile)
	if err != nil {
		log.Println("user:", user_id, "", "Failed to read markdown file", " ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	note_params.Content, err = MarkdownToHTML(w, r, images, mdContent, note_params.Path, user_id)
	if err != nil {
		writeJSONError(w, r, err, "Failed to convert Markdown to HTML (Image file is missing)", http.StatusInternalServerError)
		return
	}

	// Check if the note exists
	existingNote, err := database.Queries.GetNoteByPathAndID(r.Context(), db.GetNoteByPathAndIDParams{Path: note_params.Path, UserID: note_params.UserID, ID: note_params.ID})
	if errors.Is(err, sql.ErrNoRows) {
		writeJSONError(w, r, nil, "Note does not exist", http.StatusBadRequest)
		return
	} else if err != nil {
		writeJSONError(w, r, err, "Failed to find note", http.StatusInternalServerError)
		return
	}

	// Update the note in the database
	err = database.Queries.UpdateNote(r.Context(), note_params)
	if err != nil {
		writeJSONError(w, r, err, "Failed to update note", http.StatusInternalServerError)
		return
	}
	var ImageHashes []string
	for _, img := range images {
		ImageHashes = append(ImageHashes, img.Hash)
	}

	// Delete old image links
	err = database.Queries.UnlinkOldImagesFromNote(r.Context(), db.UnlinkOldImagesFromNoteParams{NoteID: existingNote.ID, Hashes: ImageHashes})
	if err != nil {
		writeJSONError(w, r, err, "Failed to delete old images", http.StatusInternalServerError)
		return
	}
	// Link note and images
	for _, img := range images {
		_, err = database.Queries.GetNoteImage(r.Context(), db.GetNoteImageParams{ImageID: img.Id, NoteID: existingNote.ID})
		if err == nil {
			continue
		} else if !errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, r, err, "Failed to get note, image", http.StatusInternalServerError)
			return
		}
		err := database.Queries.LinkImageToNote(r.Context(), db.LinkImageToNoteParams{ImageID: img.Id, NoteID: existingNote.ID})
		if err != nil {
			log.Println("Params: ", db.LinkImageToNoteParams{ImageID: img.Id, NoteID: existingNote.ID})
			writeJSONError(w, r, err, "Failed to link note and image", http.StatusInternalServerError)
			return
		}
	}

	// Return JSON Response
	log.Println("user:", user_id, "", "Update successful")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Update successful",
		"markdown_path": note_params.Path,
		"saved_note":    header.Filename,
		"saved_images": func() []string {
			var paths []string
			for _, img := range images {
				paths = append(paths, img.Path)
			}
			return paths
		}(),
	})
}
