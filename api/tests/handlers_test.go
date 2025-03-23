// TODO : Fix tests
// TODO : Fix image uploading without images + response
// TODO : Fix image updating without images + response

package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Tarat0r/Markdown-Blog/handlers"
	"github.com/Tarat0r/Markdown-Blog/middleware"
)

func TestListNotesHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/notes", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", os.Getenv("AUTHORIZATION"))

	rr := httptest.NewRecorder()
	handler := MiddlewareChain(middleware.LoggingMiddleware, middleware.AuthMiddleware)(handlers.ListNotes)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("ListNotes returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}
}

func TestCreateNoteHandler_All(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("metadata", `{"path":"/path/to/create_test"}`)
	part, _ := writer.CreateFormFile("markdown", "test.md")
	part.Write([]byte("# New Note"))
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
		t.Errorf("CreateNote returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	var result map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &result)
	if err != nil {
		t.Fatalf("Invalid JSON response: %v \n %s", err, rr.Body.String())
	}

	expectedMessage := "Upload successful"
	if result["message"] != expectedMessage {
		t.Errorf("unexpected message: got %v want %v", result["message"], expectedMessage)
	}
}

func TestUpdateNoteHandler_All(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("metadata", `{"path":"/path/to/update_test"}`)
	part, _ := writer.CreateFormFile("markdown", "update.md")
	part.Write([]byte("# Updated Note Content"))
	writer.Close()

	req, err := http.NewRequest("PUT", "/notes/123", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", os.Getenv("AUTHORIZATION"))

	rr := httptest.NewRecorder()
	handler := MiddlewareChain(middleware.LoggingMiddleware, middleware.AuthMiddleware)(handlers.UpdateNote)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("UpdateNote returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	var result map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &result)
	if err != nil {
		t.Fatalf("Invalid JSON response: %v \n %s", err, rr.Body.String())
	}

	expectedMessage := "Note updated successfully"
	if result["message"] != expectedMessage {
		t.Errorf("unexpected message: got %v want %v", result["message"], expectedMessage)
	}
}

func TestDeleteNoteHandler_All(t *testing.T) {
	req, err := http.NewRequest("DELETE", "/notes/123", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", os.Getenv("AUTHORIZATION"))

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
		t.Errorf("unexpected message: got %v want %v", result["message"], expectedMessage)
	}
}
