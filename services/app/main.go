package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/app/api"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

//	@title		Distributed Notification Server
//	@version	0.0.1

// main starts the Distributed Notification Server: it loads environment variables, establishes a PostgreSQL connection, initializes the API router, and listens for HTTP requests on :8080.
// It logs fatal errors on startup failures and ensures the database connection and context are cleaned up on exit.
func main() {
	godotenv.Load()

	var DATABASE_URL = os.Getenv("DATABASE_URL")
	if DATABASE_URL == "" {
		log.Fatal("failed to get database url")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, DATABASE_URL)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	api := api.Api()
	log.Println("server starting on :8080")
	if err := http.ListenAndServe(":8080", api); err != nil {
		log.Fatal(err)
	}
}