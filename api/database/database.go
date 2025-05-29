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
	const maxRetries = 5
	const retryInterval = 2 * time.Second

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	var err error
	for i := 1; i <= maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		DBPool, err = pgxpool.New(ctx, dbURL)
		if err == nil {
			err = DBPool.Ping(ctx)
			if err == nil {
				cancel()
				break // successful connection
			}
			DBPool.Close()
		}

		cancel()
		log.Printf("Attempt %d: Failed to connect to PostgreSQL: %v", i, err)
		time.Sleep(retryInterval)
	}

	if err != nil {
		log.Fatalf("Could not connect to PostgreSQL after %d attempts: %v", maxRetries, err)
	}

	Queries = db.New(DBPool)
	log.Println("Connected to PostgreSQL using pgxpool!")
}

// Close the database connection pool
func CloseDB() {
	DBPool.Close()
	log.Println("Database connection pool closed")
}

// RunMigrations executes the SQL file to set up the database schema
func RunMigrations(filePath string) {
	// filePath := "../database/markdown_blog.sql" // Path to the SQL file
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
