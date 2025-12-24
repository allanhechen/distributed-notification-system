package main

import (
	"log/slog"

	"github.com/allanhechen/distributed-notification-system/utils"
)

func main() {
	config := utils.LoadConfig()

	f := utils.ConfigureLogger(config)
	conn := utils.ConfigureDatabase(config)
	if f != nil {
		defer f.Close()
	}
	defer conn.Close()

	slog.Info("worker started")
}
