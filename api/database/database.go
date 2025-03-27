package database

import (
	"context"
	"log"
	"os"
	"time"

	db "github.com/Tarat0r/Markdown-Blog/database/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

var DBPool *pgxpool.Pool
var Queries *db.Queries

// Initialize the connection pool
func ConnectDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Load database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	// Create connection pool
	var err error
	DBPool, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	// Verify connection
	err = DBPool.Ping(ctx)
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize SQLC Queries
	Queries = db.New(DBPool)

	log.Println("Connected to PostgreSQL using pgxpool!")
}

// Close the database connection pool
func CloseDB() {
	DBPool.Close()
	log.Println("Database connection pool closed")
}

// RunMigrations executes the SQL file to set up the database schema
func RunMigrations() {
	filePath := "../database/markdown_blog.sql" // Path to the SQL file
	sqlFile, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read SQL file: %v", err)
	}

	conn, err := DBPool.Acquire(context.Background())
	if err != nil {
		log.Fatalf("Failed to acquire database connection: %v", err)
	}
	defer conn.Release()

	_, err = conn.Exec(context.Background(), string(sqlFile))
	if err != nil {
		log.Fatalf("Failed to execute migrations: %v", err)
	}

	log.Println("Database migrations executed successfully!")
}
