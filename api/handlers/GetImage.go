package handlers

import (
	"io"
	"net/http"
	"os"

	"github.com/Tarat0r/Markdown-Blog/database"
	db "github.com/Tarat0r/Markdown-Blog/database/sqlc"
)

func GetImage(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	contextUserID := r.Context().Value("contextUserID").(int32)
	// if !ok {
	// 	writeJSONError(w, r, nil, "Unauthorized", http.StatusUnauthorized)
	// 	return
	// }

	// Get image hash from URL
	imageHash := r.PathValue("ImageHash")

	_, err := database.Queries.UserCanAccessImageByHash(r.Context(), db.UserCanAccessImageByHashParams{UserID: contextUserID, Hash: imageHash})
	if err != nil {
		// writeJSONError(w, r, err, "Unauthorized", http.StatusUnauthorized)
		return
	}

	imagePath := os.Getenv("STATIC_PATH") + "/" + imageHash
	file, err := os.Open(imagePath)
	if err != nil {
		// http.Error(w, "Image not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Detect content type from first 512 bytes
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	contentType := http.DetectContentType(buf[:n])

	// Rewind file to start
	file.Seek(0, 0)

	// Set correct headers
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)

	// Stream the file to the response
	io.Copy(w, file)
}
