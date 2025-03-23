package main

import (
	"bytes"
	"encoding/json"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Tarat0r/Markdown-Blog/database"
	"github.com/Tarat0r/Markdown-Blog/handlers"
	"github.com/Tarat0r/Markdown-Blog/middleware"
	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize the database connection
	database.ConnectDB()
	defer database.CloseDB()

	// Run the tests
	m.Run()
}

func TestMainFunction(t *testing.T) {
	// Create a new HTTP request
	req, err := http.NewRequest("GET", "/notes", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Get the Authorization header value from the environment
	authHeader := os.Getenv("AUTHORIZATION")
	if authHeader == "" {
		t.Fatal("AUTHORIZATION environment variable is not set")
	}

	// Add the Authorization header
	req.Header.Set("Authorization", authHeader)

	// Create a new HTTP response recorder
	rr := httptest.NewRecorder()

	// Define your middleware chain
	middlewareChain := MiddlewareChain(middleware.LoggingMiddleware, middleware.AuthMiddleware)

	// Create a handler function
	handler := middlewareChain(handlers.ListNotes)

	// Serve the HTTP request
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	expected := `[{"id":32,"path":"/path/to/note23","hash":"6248ac82e9517029667713ea9b20643fd90fcccd3de1df3e31fd8e2803655e61"},{"id":4,"path":"test/tes3.md","hash":"no_hash2"}]`
	var expectedNotes, actualNotes []map[string]interface{}
	if err := json.Unmarshal([]byte(expected), &expectedNotes); err != nil {
		t.Fatalf("Error unmarshalling expected response: %v", err)
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &actualNotes); err != nil {
		t.Fatalf("Error unmarshalling actual response: %v", err)
	}
	if !equal(expectedNotes, actualNotes) {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestMiddlewareChain(t *testing.T) {
	// Create a new HTTP request
	req, err := http.NewRequest("GET", "/notes", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Get the Authorization header value from the environment
	authHeader := os.Getenv("AUTHORIZATION")
	if authHeader == "" {
		t.Fatal("AUTHORIZATION environment variable is not set")
	}

	// Add the Authorization header
	req.Header.Set("Authorization", authHeader)

	// Create a new HTTP response recorder
	rr := httptest.NewRecorder()

	// Define your middleware chain
	middlewareChain := MiddlewareChain(middleware.LoggingMiddleware, middleware.AuthMiddleware)

	// Create a handler function
	handler := middlewareChain(handlers.ListNotes)

	// Serve the HTTP request
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	expected := `[{"id":32,"path":"/path/to/note23","hash":"6248ac82e9517029667713ea9b20643fd90fcccd3de1df3e31fd8e2803655e61"},{"id":4,"path":"test/tes3.md","hash":"no_hash2"}]`
	var expectedNotes, actualNotes []map[string]interface{}
	if err := json.Unmarshal([]byte(expected), &expectedNotes); err != nil {
		t.Fatalf("Error unmarshalling expected response: %v", err)
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &actualNotes); err != nil {
		t.Fatalf("Error unmarshalling actual response: %v", err)
	}
	if !equal(expectedNotes, actualNotes) {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestCreateNoteHandler(t *testing.T) {
	// Create a new multipart form request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("metadata", `{"path":"/path/to/newnote"}`)
	part, _ := writer.CreateFormFile("markdown", "test.md")
	part.Write([]byte("# Test Note"))
	writer.Close()

	req, err := http.NewRequest("POST", "/notes", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Get the Authorization header value from the environment
	authHeader := os.Getenv("AUTHORIZATION")
	if authHeader == "" {
		t.Fatal("AUTHORIZATION environment variable is not set")
	}

	// Add the Authorization header
	req.Header.Set("Authorization", authHeader)

	// Create a new HTTP response recorder
	rr := httptest.NewRecorder()

	// Define your middleware chain
	middlewareChain := MiddlewareChain(middleware.LoggingMiddleware, middleware.AuthMiddleware)

	// Create a handler function
	handler := middlewareChain(handlers.CreateNote)

	// Serve the HTTP request
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	var result map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &result)
	if err != nil {
		t.Fatalf("Invalid JSON response: %v \n %s", err, rr.Body.String())
	}

	expectedMessage := "Upload successful"
	if result["message"] != expectedMessage {
		t.Errorf("unexpected message: got %v want %v", result["message"], expectedMessage)
	}

	expectedMarkdownPath := "/path/to/note23"
	if result["markdown_path"] != expectedMarkdownPath {
		t.Errorf("unexpected markdown_path: got %v want %v", result["markdown_path"], expectedMarkdownPath)
	}

	expectedSavedImages := []string{"/path/to/schema3.png", "/path/to/schema2.png", "/path/to/schema.png"}
	expectedSavedImagesInterface := make([]interface{}, len(expectedSavedImages))
	for i, v := range expectedSavedImages {
		expectedSavedImagesInterface[i] = v
	}
	if !equalStringSlices(result["saved_images"].([]interface{}), expectedSavedImagesInterface) {
		t.Errorf("unexpected saved_images: got %v want %v", result["saved_images"], expectedSavedImages)
	}

	expectedSavedNote := "note_1.md"
	if result["saved_note"] != expectedSavedNote {
		t.Errorf("unexpected saved_note: got %v want %v", result["saved_note"], expectedSavedNote)
	}
}

func TestUpdateNoteHandler(t *testing.T) {
	// Create a new multipart form request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("metadata", `{"path":"/path/to/existingnote"}`)
	part, _ := writer.CreateFormFile("markdown", "test.md")
	part.Write([]byte("# Updated Test Note"))
	writer.Close()

	req, err := http.NewRequest("PUT", "/notes/1", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Get the Authorization header value from the environment
	authHeader := os.Getenv("AUTHORIZATION")
	if authHeader == "" {
		t.Fatal("AUTHORIZATION environment variable is not set")
	}

	// Add the Authorization header
	req.Header.Set("Authorization", authHeader)

	// Create a new HTTP response recorder
	rr := httptest.NewRecorder()

	// Define your middleware chain
	middlewareChain := MiddlewareChain(middleware.LoggingMiddleware, middleware.AuthMiddleware)

	// Create a handler function
	handler := middlewareChain(handlers.UpdateNote)

	// Serve the HTTP request
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	expected := `{"message":"Note updated successfully"}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestDeleteNoteHandler(t *testing.T) {
	// Create a new HTTP request
	req, err := http.NewRequest("DELETE", "/notes/1", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Get the Authorization header value from the environment
	authHeader := os.Getenv("AUTHORIZATION")
	if authHeader == "" {
		t.Fatal("AUTHORIZATION environment variable is not set")
	}

	// Add the Authorization header
	req.Header.Set("Authorization", authHeader)

	// Create a new HTTP response recorder
	rr := httptest.NewRecorder()

	// Define your middleware chain
	middlewareChain := MiddlewareChain(middleware.LoggingMiddleware, middleware.AuthMiddleware)

	// Create a handler function
	handler := middlewareChain(handlers.DeleteNote)

	// Serve the HTTP request
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	expected := `{"message":"Note deleted successfully"}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
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

// Helper function to compare two slices of maps
func equal(a, b []map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		for k, v := range a[i] {
			if b[i][k] != v {
				return false
			}
		}
	}
	return true
}

// Helper function to compare two slices of strings
func equalStringSlices(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
