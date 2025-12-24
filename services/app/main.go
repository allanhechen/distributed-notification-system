package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/allanhechen/distributed-notification-system/services/app/api"
	"github.com/allanhechen/distributed-notification-system/services/app/internal"
	"github.com/allanhechen/distributed-notification-system/services/app/internal/db"
	"github.com/allanhechen/distributed-notification-system/services/app/internal/repository"
	"github.com/allanhechen/distributed-notification-system/services/app/internal/services"
	"github.com/allanhechen/distributed-notification-system/utils"
)

// @title		Distributed Notification Server
// @version	0.0.1
func main() {
	config := utils.LoadConfig()

	f := utils.ConfigureLogger(config)
	conn := utils.ConfigureDatabase(config)
	if f != nil {
		defer f.Close()
	}
	defer conn.Close()

	idempotencyRepo := repository.NewIdempotencyRepo(conn)
	idempotencyService := services.NewIdempotencyService(idempotencyRepo)
	idempotentRequestHandler := api.NewIdempotentRequestHandler(idempotencyService)
	idempotencyLayer := internal.NewConcreteIdempotencyLayer(conn, func(q db.Querier) repository.IdempotencyLayerRepo {
		return repository.NewIdempotencyLayerRepo(q)
	})

	apiHandler := api.Api(idempotentRequestHandler, idempotencyLayer)
	slog.Info("server starting on :8080")
	if err := http.ListenAndServe(":8080", apiHandler); err != nil {
		slog.Error("failed to start HTTP server", "error", err)
		os.Exit(1)
	}
}
