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

	var databaseUrl, urlPresent = os.LookupEnv("DATABASE_URL")
	if !urlPresent {
		log.Fatal("failed to get database url")
	}

	var logLevel, levelPresent = os.LookupEnv("LOG_LEVEL")
	if !levelPresent {
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

func configureLogger(config *Config) *os.File {
	serverInstance := uuid.New()
	var logger *slog.Logger
	var fileHandler *os.File

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
		fileHandler = f

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
	return fileHandler
}

func configureDatabase(config *Config) (*pgx.Conn, context.Context) {
	connCtx, connCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer connCancel()

	conn, err := pgx.Connect(connCtx, config.databaseUrl)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	appCtx := context.Background()
	return conn, appCtx
}

// @title		Distributed Notification Server
// @version	0.0.1
func main() {
	config := loadConfig()

	f := configureLogger(config)
	conn, ctx := configureDatabase(config)
	defer f.Close()
	defer conn.Close(ctx)

	defer func() {
		if r := recover(); r != nil {
			f.Close()
			conn.Close(ctx)
			slog.Error("server panicked, cleanly closing connections")
			panic(r)
		}
	}()

	apiHandler := api.Api()
	slog.Info("server starting on :8080")
	if err := http.ListenAndServe(":8080", apiHandler); err != nil {
		slog.Error("failed to start HTTP server", "error", err)
		os.Exit(1)
	}
}
