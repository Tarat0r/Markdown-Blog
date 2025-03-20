package handlers

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/Tarat0r/Markdown-Blog/database"
	db "github.com/Tarat0r/Markdown-Blog/database/sqlc"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
)

// JSON Metadata Struct
type UploadRequest struct {
	Path   string        `json:"path"`
	Images []ImageUpload `json:"images"`
}

// ImageUpload Struct
type ImageUpload struct {
	Path string `json:"path"`
}

// Image Struct
type Image struct {
	Id   int32
	Path string
	File multipart.File
	Hash string
}

// ComputeSHA256Hash generates a SHA-256 hash of a file
func ComputeSHA256Hash(file multipart.File) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func CreateNote(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	user_id, ok := r.Context().Value("user_id").(int32)
	if !ok {
		writeJSONError(w, r, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// File Upload Handler
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
	images := ImageUploadHandler(w, r, req, user_id)
	if images == nil {
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

	var note_params db.CreateNoteParams

	note_params.Path = req.Path
	note_params.UserID = user_id
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

	fmt.Println("Note hash:", note_params.Hash)
	mdContent, err := io.ReadAll(mdFile)
	// fmt.Println("Note", string(mdContent))

	note_params.Content = MarkdownToHTML(w, r, images, mdContent)

	log.Println(note_params.Content)

	//----------------------------------------------------------------

	// Return JSON Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Upload successful",
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

//-*-*-*-***-*-**-*--*-**--*--*-*-*-*-**--*-*-*-*-*-**--*-*-**-*--*

// ConvertMarkdown processes Markdown and applies AST modifications
func MarkdownToHTML(w http.ResponseWriter, r *http.Request, img []Image, md []byte) string {
	// Create a new Markdown parser
	gm := goldmark.New(
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithHardWraps()),
	)

	// Parse the Markdown into an AST (Abstract Syntax Tree)
	reader := text.NewReader(md)
	doc := gm.Parser().Parse(reader)

	// Call the separate function to modify AST
	// modifyAST(doc)
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		switch node := n.(type) {
		// TODO check if the image links are the same count like the image files
		// Modify image links (add CDN prefix)
		case *ast.Image:
			if entering {
				oldSrc := string(node.Destination)
				fmt.Println(oldSrc)
				newSrc := "http://localhost/static/" + img[0].Hash // TODO make it env
				node.Destination = []byte(newSrc)                  // TODO if is http image - don't change
			}

		// Modify Markdown links (.md → .html)
		case *ast.Link:
			if entering {
				oldHref := string(node.Destination)
				if strings.HasSuffix(oldHref, ".md") {
					newHref := strings.TrimSuffix(oldHref, ".md") + ".html"
					node.Destination = []byte(newHref)
				}
			}
		}
		return ast.WalkContinue, nil
	})

	// Render the modified AST back into HTML
	var buf bytes.Buffer
	if err := gm.Renderer().Render(&buf, md, doc); err != nil {
		log.Fatal(err)
	}
	return buf.String()
}

// ImageUploadHandler handles image uploads
func ImageUploadHandler(w http.ResponseWriter, r *http.Request, req UploadRequest, user_id int32) []Image {
	var img Image
	var images []Image // Slice to hold images
	var err error

	log.Println("req:", req)
	uploadedImages := r.MultipartForm.File["image"]

	// Check if the number of uploaded images matches the number of image paths in JSON metadata
	if len(uploadedImages) != len(req.Images) {
		writeJSONError(w, r, nil, "The number of uploaded images does not match the number of image paths in JSON metadata", http.StatusBadRequest)
		return nil
	}

	for i, path := range req.Images {
		img.Path = path.Path
		img.File, err = uploadedImages[i].Open()
		if err != nil {
			http.Error(w, "File upload error", http.StatusBadRequest)
			return nil
		}
		defer img.File.Close()

		// Check MIME type
		buffer := make([]byte, 512)
		if _, err := img.File.Read(buffer); err != nil {
			log.Println("user:", user_id, "", "Failed to read file", " ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return nil
		}
		if _, err := img.File.Seek(0, io.SeekStart); err != nil {
			log.Println("user:", user_id, "", "Failed to reset file pointer", " ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return nil
		}
		mimeType := http.DetectContentType(buffer)
		if !strings.HasPrefix(mimeType, "image/") {
			writeJSONError(w, r, err, "Invalid image file type", http.StatusBadRequest)
			return nil
		}

		// Compute hash of image file
		img.Hash, err = ComputeSHA256Hash(img.File)
		if err != nil {
			log.Println("user:", user_id, "", "Failed to compute image file hash", " ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return nil
		}

		//Check if image already exists
		_, err := database.Queries.GetImageByHash(r.Context(), img.Hash)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// Save image if it does not exist
				savePath := os.Getenv("STATIC_PATH") + "/" + img.Hash
				outFile, err := os.Create(savePath)
				if err != nil {
					writeJSONError(w, r, err, "Failed to save image", http.StatusInternalServerError)
					return nil
				}
				defer outFile.Close()
				io.Copy(outFile, img.File)

				// Add to database
				img.Id, err = database.Queries.UploadImage(r.Context(), img.Hash)
				if err != nil {
					log.Println("user:", user_id, "", "Failed to save image to database", " ", err)
					w.WriteHeader(http.StatusInternalServerError)
					return nil
				}
				log.Println("user:", user_id, "", "Image saved to:", savePath)
			} else {
				log.Println("user:", user_id, "", "Failed to get hash from database", " ", err)
				w.WriteHeader(http.StatusInternalServerError)
				return nil
			}
		}
		images = append(images, img)
	}
	return images
}
