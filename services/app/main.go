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

// @title		Distributed Notification Server
// @version	0.0.1
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

	apiHandler := api.Api()
	log.Println("server starting on :8080")
	if err := http.ListenAndServe(":8080", apiHandler); err != nil {
		log.Fatal(err)
	}
}
