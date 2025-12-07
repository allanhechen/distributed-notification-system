package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/allanhechen/distributed-notification-system/services/app/api"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
)

// Config loaded from env to configure application behavior
type Config struct {
	databaseUrl string
	logLevel    string
}

func loadConfig() *Config {
	godotenv.Load()

	var databaseUrl = os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		log.Fatal("failed to get database url")
	}

	var logLevel = os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		log.Fatal("failed to get log level")
	}
	logLevel = strings.ToLower(logLevel)
	if logLevel != "development" && logLevel != "production" {
		log.Fatal("log level set to an invalid value")
	}

	return &Config{
		databaseUrl,
		logLevel,
	}
}

// @title		Distributed Notification Server
// @version	0.0.1
func main() {
	config := loadConfig()
	serverInstance := uuid.New()

	var logger *slog.Logger
	switch config.logLevel {
	case "production":
		err := os.MkdirAll("logs", 0o755)
		if err != nil {
			log.Fatalf("could not create logs directory: %v", err)
		}

		now := time.Now().UTC().Format("20060102T150405Z0700")
		logName := fmt.Sprintf("server_log_%s_%s.log", now, serverInstance)
		logPath := filepath.Join("./logs", logName)
		f, err := os.Create(logPath)

		if err != nil {
			log.Fatalf("could not create log file: %v", err)
		}

		logger = slog.New(slog.NewJSONHandler(f, &slog.HandlerOptions{
			ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.Time(slog.TimeKey, a.Value.Time().UTC())
				}
				return a
			},
		}))
		logger = logger.With("service_id", serverInstance)
	case "development":
		logger = slog.New(tint.NewHandler(os.Stdout, nil))
	}
	slog.SetDefault(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, config.databaseUrl)
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
