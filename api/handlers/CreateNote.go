package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/Tarat0r/Markdown-Blog/database"
	db "github.com/Tarat0r/Markdown-Blog/database/sqlc"
	"github.com/Tarat0r/Markdown-Blog/notifications"
	obsidian "github.com/powerman/goldmark-obsidian"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/wikilink"
)

// JSON Metadata Struct
type UploadRequest struct {
	Path   string        `json:"path"`
	Images []ImageUpload `json:"images,omitempty"`
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

func CreateNote(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	contextUserID, ok := r.Context().Value("contextUserID").(int32)
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
	images, err := ImageUploadHandler(w, r, req, contextUserID)
	if err != nil {
		log.Println("user:", contextUserID, "", err)
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
	n, err := mdFile.Read(buffer)
	if err != nil {
		log.Println("user:", contextUserID, "", "Failed to read file", " ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, err := mdFile.Seek(0, io.SeekStart); err != nil {
		log.Println("user:", contextUserID, "", "Failed to reset file pointer", " ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	mimeType := http.DetectContentType(buffer)
	if !IsValidMarkdown(mimeType, n) {
		log.Println("user:", contextUserID, "", strings.TrimSpace(mimeType), " ", header.Filename)
		writeJSONError(w, r, nil, "Invalid markdown file type", http.StatusBadRequest)
		return
	}

	var noteParams db.CreateNoteParams

	noteParams.Path = req.Path
	noteParams.UserID = contextUserID
	// Compute SHA-256 Hash of Markdown file
	noteParams.Hash, err = ComputeSHA256Hash(mdFile)
	if err != nil {
		log.Println("user:", contextUserID, "", "Failed to compute markdown file hash", " ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Reset file pointer to the beginning before reading the content
	if _, err := mdFile.Seek(0, io.SeekStart); err != nil {
		log.Println("user:", contextUserID, "", "Failed to reset file pointer", " ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	mdContent, err := io.ReadAll(mdFile)
	if err != nil {
		log.Println("user:", contextUserID, "", "Failed to read markdown file", " ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	noteParams.ContentMd = string(mdContent)

	noteParams.Content, err = MarkdownToHTML(w, r, images, mdContent, noteParams.Path, contextUserID)
	if err != nil {
		writeJSONError(w, r, err, "Failed to convert Markdown to HTML (Image file is missing)", http.StatusInternalServerError)
		return
	}

	notesByPath, err := database.Queries.GetNoteByPath(r.Context(), db.GetNoteByPathParams{Path: noteParams.Path, UserID: noteParams.UserID})
	if err != nil {
		writeJSONError(w, r, err, "Failed to create note", http.StatusInternalServerError)
		return
	}
	if len(notesByPath) > 0 {
		writeJSONError(w, r, errors.New("Note already exists"), "Note already exists", http.StatusBadRequest)
		return
	}

	uploadedNote, err := database.Queries.CreateNote(r.Context(), noteParams)
	if err != nil {
		writeJSONError(w, r, err, "Failed to create note", http.StatusInternalServerError)
		return
	}

	// Link note and images
	for _, img := range images {

		_, err = database.Queries.GetNoteImage(r.Context(), db.GetNoteImageParams{ImageID: img.Id, NoteID: uploadedNote.ID})
		if err == nil {
			continue
		} else if !errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, r, err, "Failed to get note, image", http.StatusInternalServerError)
			return
		}
		err = database.Queries.LinkImageToNote(r.Context(), db.LinkImageToNoteParams{ImageID: img.Id, NoteID: uploadedNote.ID})
		if err != nil {
			log.Println("Params: ", db.LinkImageToNoteParams{ImageID: img.Id, NoteID: uploadedNote.ID})
			writeJSONError(w, r, err, "Failed to link note and image", http.StatusInternalServerError)
			return
		}
	}

	// Return JSON Response
	// log.Println("user:", contextUserID, "", "Note created successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Upload successful",
		"note_id":       uploadedNote.ID,
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

	// внутри handler'а или бизнес-логики
	notifications.NotifyTelegram("📝 Новая заметка создана! \n Пользователь: " + strconv.Itoa(int(contextUserID)) + "\n ID заметки: " + strconv.Itoa(int(uploadedNote.ID)))
}

//-*-*-*-***-*-**-*--*-**--*--*-*-*-*-**--*-*-*-*-*-**--*-*-**-*--*

// CustomResolver is a wikilink.Resolver that returns the target as-is (no .html).
type CustomResolver struct{}

func (r CustomResolver) ResolveWikilink(info *wikilink.Node) ([]byte, error) {
	// info.Target is the contents of [[link]]
	return info.Target, nil // no .html suffix, no transformation
}

// ConvertMarkdown processes Markdown and applies AST modifications
func MarkdownToHTML(w http.ResponseWriter, r *http.Request, img []Image, md []byte, notePath string, userID int32) (string, error) {
	// Create a new Markdown parser
	gm := goldmark.New(
		goldmark.WithExtensions(
			obsidian.NewPlugTasks(),
			obsidian.NewObsidian(),
			&wikilink.Extender{
				Resolver: CustomResolver{},
			},
		),
	)
	// Parse the Markdown into an AST (Abstract Syntax Tree)
	reader := text.NewReader(md)
	doc := gm.Parser().Parse(reader)

	// Call the separate function to modify AST
	// modifyAST(doc) // Placeholder for future AST modifications, currently not in use.

	imageCount := 0
	rx := regexp.MustCompile(`(?i)^https?://[^\s/$.?#].[^\s]*$`)
	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		switch node := n.(type) {
		// Modify image links (add CDN prefix)
		case *ast.Image:
			if entering {
				oldSrc := string(node.Destination)
				if !rx.MatchString(oldSrc) {
					if imageCount >= len(img) {
						return ast.WalkStop, errors.New("Image count mismatch: number of images in metadata does not match the number of uploaded files")
					}

					staticPath := os.Getenv("STATIC_PATH")
					if staticPath == "" {
						return ast.WalkStop, errors.New("STATIC_PATH environment variable is not set")
					}
					newSrc := staticPath + "/" + img[imageCount].Hash

					node.Destination = []byte(newSrc)
					imageCount++
				}
			}

			// Obsidian Links
		case *wikilink.Node:
			if entering && !rx.MatchString(string(node.Target)) {
				if node.Embed {
					if imageCount >= len(img) {
						return ast.WalkStop, errors.New("Image count does not match")
					}
					newSrc := os.Getenv("STATIC_PATH") + "/" + img[imageCount].Hash

					imageCount++
					node.Target = []byte(newSrc)
				} else {
					oldLink := string(node.Target)
					// TODO Change the link format, if needed
					node.Target = []byte(oldLink)

				}
			}

		}
		return ast.WalkContinue, nil

	})

	if err != nil {
		return "", err
	}

	// Render the modified AST back into HTML
	var buf bytes.Buffer
	if err := gm.Renderer().Render(&buf, md, doc); err != nil {
		// log.Println("user:", userID, "", "Failed to render HTML", " ", err)
		return "", err
	}
	return buf.String(), nil
}

// ImageUploadHandler handles image uploads
func ImageUploadHandler(w http.ResponseWriter, r *http.Request, req UploadRequest, contextUserID int32) ([]Image, error) {
	var img Image
	var images []Image // Slice to hold images
	var err error

	uploadedImages := r.MultipartForm.File["image"]

	// Check if images are uploaded
	if len(uploadedImages) == 0 {
		return nil, nil
	}

	// Check if the number of uploaded images matches the number of image paths in JSON metadata
	if len(uploadedImages) != len(req.Images) {
		return nil, errors.New("The number of uploaded images does not match the number of image paths in JSON metadata")
	}

	for i, path := range req.Images {
		img.Path = path.Path
		img.File, err = uploadedImages[i].Open()
		if err != nil {
			return nil, errors.New("File upload error")
		}
		defer img.File.Close()

		// Check MIME type
		buffer := make([]byte, 512)
		if _, err := img.File.Read(buffer); err != nil {
			return nil, errors.New("Failed to read image file")
		}
		if _, err := img.File.Seek(0, io.SeekStart); err != nil {
			return nil, errors.New("Failed to reset file pointer")
		}
		mimeType := http.DetectContentType(buffer)
		if !strings.HasPrefix(mimeType, "image/") {
			return nil, errors.New("Invalid image file type")
		}

		// Compute hash of image file
		img.Hash, err = ComputeSHA256Hash(img.File)
		if err != nil {
			return nil, errors.New("Failed to compute image file hash")
		}

		//Check if image already exists
		imgFromDB, err := database.Queries.GetImageByHash(r.Context(), img.Hash)
		img.Id = imgFromDB.ID
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// Save image if it does not exist
				// savePath := os.Getenv("STATIC_PATH") + "/" + img.Hash

				staticDir := getStaticPath()
				if staticDir == "" {
					return nil, errors.New("STATIC_PATH env var is not set")
				}

				if err := os.MkdirAll(staticDir, 0o755); err != nil {
					return nil, errors.New("failed to create STATIC_PATH " + staticDir + ": " + err.Error())
				}
				savePath := filepath.Join(staticDir, img.Hash)

				// Re-open the file to read it again from start
				imgReader, err := uploadedImages[i].Open()
				if err != nil {
					return nil, errors.New("Failed to re-open image")
				}
				defer imgReader.Close()

				outFile, err := os.Create(savePath)
				if err != nil {
					return nil, errors.New("Failed to save image")
				}
				defer outFile.Close()

				_, err = io.Copy(outFile, imgReader)
				if err != nil {
					return nil, errors.New("Failed to write image to file")
				}

				// Add to database
				img.Id, err = database.Queries.UploadImage(r.Context(), img.Hash)
				if err != nil {
					return nil, errors.New("Failed to save image to database")
				}
				log.Println("user:", contextUserID, "", "Image saved to:", savePath)
			} else {
				return nil, errors.New("Failed to get hash from database")
			}
		}
		images = append(images, img)
	}
	return images, nil
}

func getStaticPath() string {
	return os.Getenv("STATIC_PATH")
}

func IsValidMarkdown(mimeType string, n int) bool {
	return strings.HasPrefix(mimeType, "text/plain") ||
		(mimeType == "application/octet-stream" && n < 512)
}
