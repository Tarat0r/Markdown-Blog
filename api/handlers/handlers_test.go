// TODO : Fix tests
// TODO : Fix image uploading without images + response
// TODO : Fix image updating without images + response

package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"testing"

	"github.com/Tarat0r/Markdown-Blog/database"
	db "github.com/Tarat0r/Markdown-Blog/database/sqlc"
	"github.com/Tarat0r/Markdown-Blog/handlers"
	"github.com/Tarat0r/Markdown-Blog/middleware"
)

var testNoteID string

func TestMain(m *testing.M) {
	// Load environment variables from .env file
	// err := godotenv.Load("./.env")
	// if err != nil {
	// 	log.Fatalf("Error loading .env file: %v", err)
	// }

	if os.Getenv("DATABASE_URL") == "" {
		os.Setenv("DATABASE_URL", os.Getenv("DATABASE_URL"))
	}

	// Initialize the database connection
	database.ConnectDB()
	defer database.CloseDB()

	//Add test token to the database
	err := database.Queries.SetTestToken(context.Background(), db.SetTestTokenParams{ApiToken: os.Getenv("AUTHORIZATION"), Name: "TEST", Email: "test@test.com"})
	if err != nil {
		log.Fatalf("Failed to set test token: %v", err)
	}
	// Run the tests
	m.Run()
}

func TestListNotesHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/notes", nil)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(os.Getenv("AUTHORIZATION"))
	req.Header.Set("Authorization", os.Getenv("AUTHORIZATION"))

	rr := httptest.NewRecorder()
	handler := MiddlewareChain(middleware.LoggingMiddleware, middleware.AuthMiddleware)(handlers.ListNotes)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		log.Println(rr)
		t.Errorf("ListNotes returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}
}

func TestCreateNoteNoImg(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Set metadata with only path, no images
	metadata := `{"path":"test/dir1/TEST.md"}`
	_ = writer.WriteField("metadata", metadata)

	// Create markdown file part
	f, err := os.Open("./test.md")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	partHeader := textproto.MIMEHeader{}
	partHeader.Set("Content-Disposition", `form-data; name="markdown"; filename="test.md"`)
	partHeader.Set("Content-Type", "text/markdown")
	part, err := writer.CreatePart(partHeader)
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.Copy(part, f)
	if err != nil {
		t.Fatal(err)
	}

	writer.Close()

	req, err := http.NewRequest("POST", "/notes", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", os.Getenv("AUTHORIZATION"))

	rr := httptest.NewRecorder()
	handler := MiddlewareChain(middleware.LoggingMiddleware, middleware.AuthMiddleware)(handlers.CreateNote)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		log.Println(rr)
		t.Fatalf("CreateNote returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	var result map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &result)
	if err != nil {
		t.Fatalf("Invalid JSON response: %v\n%s", err, rr.Body.String())
	}

	expected := map[string]interface{}{
		"message":       "Upload successful",
		"markdown_path": "test/dir1/TEST.md",
		"saved_note":    "test.md",
		"saved_images":  nil,
	}

	noteIDVal, exists := result["note_id"]
	if !exists {
		t.Fatal("missing note_id in response")
	}

	noteIDFloat, ok := noteIDVal.(float64)
	if !ok {
		t.Fatalf("note_id is not a number: got %T", noteIDVal)
	}

	testNoteID = fmt.Sprintf("%d", int(noteIDFloat)) // Convert float64 -> int -> string

	for key, expectedVal := range expected {
		actualVal, exists := result[key]
		if !exists {
			t.Errorf("missing key in response: %v", key)
			continue
		}

		// nil check
		if expectedVal == nil {
			if actualVal != nil {
				t.Errorf("expected nil for key %s but got %v", key, actualVal)
			}
			continue
		}

		if actualVal != expectedVal {
			t.Errorf("unexpected value for key %s: got %v want %v", key, actualVal, expectedVal)
		}
	}
}

func TestUpdateNote_NoImages(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Metadata to update an existing note
	metadata := `{"path":"test/dir1/TEST.md"}`
	_ = writer.WriteField("metadata", metadata)

	// Use real markdown file
	f, err := os.Open("./test.md")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	partHeader := textproto.MIMEHeader{}
	partHeader.Set("Content-Disposition", `form-data; name="markdown"; filename="test.md"`)
	partHeader.Set("Content-Type", "text/markdown")
	part, err := writer.CreatePart(partHeader)
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.Copy(part, f)
	if err != nil {
		t.Fatal(err)
	}

	writer.Close()

	req, err := http.NewRequest("PUT", "/notes/"+testNoteID, body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", os.Getenv("AUTHORIZATION"))
	req.SetPathValue("NoteID", testNoteID)

	rr := httptest.NewRecorder()
	handler := MiddlewareChain(middleware.LoggingMiddleware, middleware.AuthMiddleware)(handlers.UpdateNote)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		log.Println(rr)
		t.Fatalf("UpdateNote returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	var result map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &result)
	if err != nil {
		t.Fatalf("Invalid JSON response: %v\n%s", err, rr.Body.String())
	}

	expected := map[string]interface{}{
		"message":       "Update successful",
		"markdown_path": "test/dir1/TEST.md",
		"saved_note":    "test.md",
		"saved_images":  nil,
	}

	for key, expectedVal := range expected {
		actualVal, exists := result[key]
		if !exists {
			t.Errorf("missing key in response: %v", key)
			continue
		}

		if expectedVal == nil {
			if actualVal != nil {
				t.Errorf("expected nil for key %s but got %v", key, actualVal)
			}
			continue
		}

		if actualVal != expectedVal {
			t.Errorf("unexpected value for key %s: got %v want %v", key, actualVal, expectedVal)
		}
	}
}

func TestDeleteNoteHandler_All(t *testing.T) {
	req, err := http.NewRequest("DELETE", "/notes/"+testNoteID, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", os.Getenv("AUTHORIZATION"))
	req.SetPathValue("NoteID", testNoteID)

	rr := httptest.NewRecorder()
	handler := MiddlewareChain(middleware.LoggingMiddleware, middleware.AuthMiddleware)(handlers.DeleteNote)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("DeleteNote returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	var result map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &result)
	if err != nil {
		t.Fatalf("Invalid JSON response: %v \n %s", err, rr.Body.String())
	}

	expectedMessage := "Note deleted successfully"
	if result["message"] != expectedMessage {
		log.Println(rr)
		t.Errorf("unexpected message: got %v want %v", result["message"], expectedMessage)
	}
}

// Middleware type definition
type Middleware func(http.HandlerFunc) http.HandlerFunc

func MiddlewareChain(middlewares ...Middleware) Middleware {
	return func(handler http.HandlerFunc) http.HandlerFunc {
		for _, mw := range middlewares {
			handler = mw(handler)
		}
		return handler
	}
}
