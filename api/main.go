package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Tarat0r/Markdown-Blog/handlers"
	"github.com/Tarat0r/Markdown-Blog/middleware"

	"github.com/Tarat0r/Markdown-Blog/database"

	_ "github.com/joho/godotenv/autoload" // Auto-load .env file
)

func main() {
	// database.ConnectDatabase()
	database.ConnectDB()
	defer database.CloseDB() // Close connection pool on exit

	// Define your middleware chain
	middlewareChain := MiddlewareChain(middleware.LoggingMiddleware, middleware.AuthMiddleware)

	http.HandleFunc("GET /notes", middlewareChain(handlers.ListNotes))
	http.HandleFunc("GET /notes/{NoteID}", middlewareChain(handlers.GetNote))

	http.HandleFunc("POST /notes", middlewareChain(handlers.CreateNote))

	http.HandleFunc("PUT /notes/{NoteID}", middlewareChain(handlers.UpdateNote))

	http.HandleFunc("DELETE /notes/{NoteID}", middlewareChain(handlers.DeleteNote))

	// fs := http.FileServer(http.Dir("frontend/static/"))
	// http.Handle("/static/", http.StripPrefix("/static/", fs))

	fmt.Println("Server is working on http://localhost:8080")
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
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
