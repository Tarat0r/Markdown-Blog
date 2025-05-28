package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Tarat0r/Markdown-Blog/database"
	db "github.com/Tarat0r/Markdown-Blog/database/sqlc"
	"github.com/Tarat0r/Markdown-Blog/notifications"
)

func UpdateNote(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	contextUserID := r.Context().Value("contextUserID").(int32)
	// if !ok {
	// 	// writeJSONError(w, r, nil, "Unauthorized", http.StatusUnauthorized)
	// 	return
	// }

	// Parse multipart form (Max 50MB)
	err := r.ParseMultipartForm(50 << 20)
	// if err != nil {
	// 	http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
	// 	return
	// }

	// Extract JSON metadata
	var req UploadRequest
	metadata := r.FormValue("metadata")
	if metadata == "" {
		// writeJSONError(w, r, nil, "Metadata is required", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal([]byte(metadata), &req); err != nil {
		// writeJSONError(w, r, err, "Invalid JSON metadata", http.StatusBadRequest)
		return
	}

	// Handle Image Uploads
	images, err := ImageUploadHandler(w, r, req, contextUserID)
	if err != nil {
		// // log.Println("user:", contextUserID, "", err)
		// writeJSONError(w, r, err, "Failed to save images", http.StatusInternalServerError)
		return
	}

	// Handle Markdown file upload
	markdownFiles := r.MultipartForm.File["markdown"]
	if len(markdownFiles) != 1 {
		// writeJSONError(w, r, nil, "Exactly one Markdown file is required", http.StatusBadRequest)
		return
	}

	mdFile, header, err := r.FormFile("markdown")
	// if err != nil {
	// writeJSONError(w, r, err, "Markdown file is required", http.StatusBadRequest)
	// return
	// }
	defer mdFile.Close()

	// Check MIME type of Markdown file
	buffer := make([]byte, 512)
	n, err := mdFile.Read(buffer)
	// if err != nil {
	// log.Println("user:", contextUserID, "", "Failed to read file", " ", err)
	// w.WriteHeader(http.StatusInternalServerError)
	// return
	// }
	if _, err := mdFile.Seek(0, io.SeekStart); err != nil {
		// log.Println("user:", contextUserID, "", "Failed to reset file pointer", " ", err)
		// w.WriteHeader(http.StatusInternalServerError)
		return
	}
	mimeType := http.DetectContentType(buffer)

	isValidMarkdown := strings.HasPrefix(mimeType, "text/plain") ||
		(mimeType == "application/octet-stream" && n < 512)

	if !isValidMarkdown {
		log.Println("user:", contextUserID, "", strings.TrimSpace(mimeType), " ", header.Filename)
		// writeJSONError(w, r, nil, "Invalid markdown file type", http.StatusBadRequest)
		return
	}
	var noteParams db.UpdateNoteParams

	noteParams.UserID = contextUserID
	noteParams.Path = req.Path
	var ok bool
	noteParams.ID, ok = GetIDFromURI(w, r, contextUserID)
	if !ok {
		return
	}
	// Compute SHA-256 Hash of Markdown file
	noteParams.Hash, err = ComputeSHA256Hash(mdFile)
	// if err != nil {
	// log.Println("user:", contextUserID, "", "Failed to compute markdown file hash", " ", err)
	// w.WriteHeader(http.StatusInternalServerError)
	// return
	// }

	// Reset file pointer to the beginning before reading the content
	if _, err := mdFile.Seek(0, io.SeekStart); err != nil {
		// log.Println("user:", contextUserID, "", "Failed to reset file pointer", " ", err)
		// w.WriteHeader(http.StatusInternalServerError)
		return
	}

	mdContent, err := io.ReadAll(mdFile)
	// if err != nil {
	// log.Println("user:", contextUserID, "", "Failed to read markdown file", " ", err)
	// w.WriteHeader(http.StatusInternalServerError)
	// return
	// }
	noteParams.ContentMd = string(mdContent)

	noteParams.Content, err = MarkdownToHTML(w, r, images, mdContent, noteParams.Path, contextUserID)
	// if err != nil {
	// writeJSONError(w, r, err, "Failed to convert Markdown to HTML (Image file is missing)", http.StatusInternalServerError)
	// return
	// }

	// Check if the note exists
	existingNote, err := database.Queries.GetNoteByPathAndID(r.Context(), db.GetNoteByPathAndIDParams{Path: noteParams.Path, UserID: noteParams.UserID, ID: noteParams.ID})
	// if errors.Is(err, sql.ErrNoRows) {
	// writeJSONError(w, r, errors.New("Note does not exist"), "Note does not exist", http.StatusBadRequest)
	// return
	// } else if err != nil {
	// writeJSONError(w, r, err, "Failed to find note", http.StatusInternalServerError)
	// return
	// }

	// Update the note in the database
	err = database.Queries.UpdateNote(r.Context(), noteParams)
	// if err != nil {
	// writeJSONError(w, r, err, "Failed to update note", http.StatusInternalServerError)
	// return
	// }
	var ImageHashes []string
	for _, img := range images {
		ImageHashes = append(ImageHashes, img.Hash)
	}

	// Delete old image links
	err = database.Queries.UnlinkOldImagesFromNote(r.Context(), db.UnlinkOldImagesFromNoteParams{NoteID: existingNote.ID, Hashes: ImageHashes})
	// if err != nil {
	// writeJSONError(w, r, err, "Failed to delete old images", http.StatusInternalServerError)
	// return
	// }
	// Link note and images
	for _, img := range images {
		_, err = database.Queries.GetNoteImage(r.Context(), db.GetNoteImageParams{ImageID: img.Id, NoteID: existingNote.ID})
		if err == nil {
			continue
		}
		// } else if !errors.Is(err, sql.ErrNoRows) {
		// writeJSONError(w, r, err, "Failed to get note, image", http.StatusInternalServerError)
		// return
		// }
		err = database.Queries.LinkImageToNote(r.Context(), db.LinkImageToNoteParams{ImageID: img.Id, NoteID: existingNote.ID})
		// if err != nil {
		// log.Println("Params: ", db.LinkImageToNoteParams{ImageID: img.Id, NoteID: existingNote.ID})
		// writeJSONError(w, r, err, "Failed to link note and image", http.StatusInternalServerError)
		// return
		// }
	}

	// Return JSON Response
	// log.Println("user:", contextUserID, "", "Update successful")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Update successful",
		"markdown_path": noteParams.Path,
		"saved_note":    header.Filename,
		"saved_images": func() []string {
			var paths []string
			for _, img := range images {
				paths = append(paths, img.Path)
			}
			return paths
		}(),
	})
	notifications.NotifyTelegram("✍🏻 Заметка редактированная! \n Пользователь: " + strconv.Itoa(int(contextUserID)) + "\n ID заметки: " + strconv.Itoa(int(noteParams.ID)))

}
