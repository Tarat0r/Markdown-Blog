package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Tarat0r/Markdown-Blog/handlers"
	"github.com/Tarat0r/Markdown-Blog/middleware"

	"github.com/Tarat0r/Markdown-Blog/database"

	_ "github.com/joho/godotenv/autoload" // Auto-load .env file
)

func main() {
	if err := StartServer(); err != nil {
		log.Fatal(err)
	}
}

func StartServer() error { // NOSONAR
	database.ConnectDB()
	defer database.CloseDB() // Close connection pool on exit

	// Run database migrations
	database.RunMigrations("../database/markdown_blog.sql")

	// Define middleware chain
	middlewareChain := MiddlewareChain(middleware.LoggingMiddleware, middleware.AuthMiddleware)

	http.HandleFunc("GET /notes", middlewareChain(handlers.ListNotes))
	http.HandleFunc("GET /notes/{NoteID}", middlewareChain(handlers.GetNote))
	http.HandleFunc("GET /images/{ImageHash}", middlewareChain(handlers.GetImage))

	http.HandleFunc("POST /notes", middlewareChain(handlers.CreateNote))

	http.HandleFunc("PUT /notes/{NoteID}", middlewareChain(handlers.UpdateNote))

	http.HandleFunc("DELETE /notes/{NoteID}", middlewareChain(handlers.DeleteNote))

	hostAddress := os.Getenv("HOST_ADDRESS")
	if hostAddress == "" {
		hostAddress = "localhost:8080"
	}

	log.Println("Server is working on http://" + hostAddress)

	return http.ListenAndServe(hostAddress, nil)
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
