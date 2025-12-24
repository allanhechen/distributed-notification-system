package utils

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/lmittmann/tint"
)

// Default timeout for database operations
var DatabaseTimeout = 24 * time.Second

// Config loaded from env to configure application behavior
type CommonConfig struct {
	databaseUrl string
	logLevel    string
}

// LoadConfig loads the configuration from the environment variables
func LoadConfig() *CommonConfig {
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

	return &CommonConfig{
		databaseUrl,
		logLevel,
	}
}

// ConfigureLogger configures slog to production or development levels as
// specified in the given config.
func ConfigureLogger(config *CommonConfig) *os.File {
	var logger *slog.Logger
	var fileHandler *os.File = nil

	switch config.logLevel {
	case "production":
		serverInstance := uuid.New()
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

// ConfigureDatabase connects to the database url specified in the config
// and returns a pooled connection
func ConfigureDatabase(config *CommonConfig) *pgxpool.Pool {
	// context only for database connections
	poolCtx, poolCancel := context.WithTimeout(context.Background(), DatabaseTimeout)
	defer poolCancel()

	pool, err := pgxpool.New(poolCtx, config.databaseUrl)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	if err := pool.Ping(poolCtx); err != nil {
		slog.Error("failed to ping database", "error", err)
		pool.Close()
		os.Exit(1)
	}

	return pool
}
