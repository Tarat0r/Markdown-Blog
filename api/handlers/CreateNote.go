package handlers

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
	Path string
	Name string
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

func CreateNote(w http.ResponseWriter, r *http.Request) { //TODO Verify file's Mime type
	user_id, ok := r.Context().Value("user_id").(int32)
	if !ok {
		writeJSONError(w, r, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// File Upload Handler
	// ✅ Parse multipart form (Max 50MB)
	err := r.ParseMultipartForm(50 << 20)
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	// ✅ Extract JSON metadata
	var req UploadRequest
	metadata := r.FormValue("metadata")
	if err := json.Unmarshal([]byte(metadata), &req); err != nil {
		writeJSONError(w, r, err, "Invalid JSON metadata", http.StatusBadRequest)
		return
	}
	fmt.Print("metadata:", metadata)
	//---------------------------------------------------------------- TO TUK SAM STIGNALI

	// TODO slice and append after for loop
	// ✅ Handle Image Uploads
	uploadedImages := r.MultipartForm.File["images"]
	savedImages := make(map[string]string) // To store matched images
	var img Image
	for _, i := range req.Images {
		filename := filepath.Base(i.Path)
		// imagePathMap[filename] = i.Path
		img.Name = filename
		img.Path = i.Path
	}

	for _, fileHeader := range uploadedImages {
		img.File, err = fileHeader.Open()
		if err != nil {
			http.Error(w, "File upload error", http.StatusBadRequest)
			return
		}
		defer img.File.Close()

		// Compute hash
		img.Hash, err = ComputeSHA256Hash(img.File)
		if err != nil {
			log.Println("user:", user_id, "", "Failed to compute image file hash", " ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fmt.Println("img struct:", img)
		fmt.Println(savedImages[""])
		// ✅ Save file using expected path
		savePath := "./static/" + img.Hash // Ensure correct mapping
		outFile, err := os.Create(savePath)
		if err != nil {
			// log.Println(err)
			// http.Error(w, "Failed to save image", http.StatusInternalServerError)
			writeJSONError(w, r, err, "Failed to save image", http.StatusInternalServerError)
			return
		}
		defer outFile.Close()
		io.Copy(outFile, img.File)

		savedImages[fileHeader.Filename] = savePath
	}

	//----------------------------------------------------------------

	// ✅ Handle Markdown file upload
	mdFile, header, err := r.FormFile("markdown")
	if err != nil {
		writeJSONError(w, r, err, "Markdown file is required", http.StatusBadRequest)
		return
	}
	defer mdFile.Close()

	var note_params db.CreateNoteParams

	// ✅ Compute SHA-256 Hash of Markdown file
	note_params.Hash, err = ComputeSHA256Hash(mdFile)
	if err != nil {
		log.Println("user:", user_id, "", "Failed to compute markdown file hash", " ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Println("hash:", note_params.Hash)

	// ✅ Save Markdown file
	mdSavePath := "./static/" + header.Filename //TODO env
	outFile, err := os.Create(mdSavePath)
	if err != nil {
		http.Error(w, "Failed to save Markdown file", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()
	io.Copy(outFile, mdFile)

	// ✅ Read Markdown file content
	mdContent, err := os.ReadFile(mdSavePath)
	if err != nil {
		http.Error(w, "Failed to read Markdown file", http.StatusInternalServerError)
		return
	}

	//----------------------------------------------------------------

	// ✅ Return JSON Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Upload successful",
		"markdown_path": mdSavePath,
		"markdown_text": string(mdContent),
		"saved_images":  savedImages,
	})
}

//-*-*-*-***-*-**-*--*-**--*--*-*-*-*-**--*-*-*-*-*-**--*-*-**-*--*

// ConvertMarkdown processes Markdown and applies AST modifications
func MarkdownToHTML(mdText string) string {
	md := []byte(mdText)

	// Create a new Markdown parser
	gm := goldmark.New(
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithHardWraps()),
	)

	// Parse the Markdown into an AST (Abstract Syntax Tree)
	reader := text.NewReader(md)
	doc := gm.Parser().Parse(reader)

	// ✅ Call the separate function to modify AST
	// modifyAST(doc)
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		switch node := n.(type) {

		// ✅ Modify image links (add CDN prefix)
		case *ast.Image:
			if entering {
				oldSrc := string(node.Destination)
				newSrc := "https://cdn.example.com" + oldSrc
				node.Destination = []byte(newSrc)
			}

		// ✅ Modify Markdown links (.md → .html)
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
