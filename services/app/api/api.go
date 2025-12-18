// Serves as the main entrypoint into the API handlers of the application
package api

import (
	"encoding/json"
	"net/http"
	"time"

	v1 "github.com/allanhechen/distributed-notification-system/services/app/api/v1"
	_ "github.com/allanhechen/distributed-notification-system/services/app/docs"
	"github.com/allanhechen/distributed-notification-system/services/app/internal"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Healthcheck endpoint
//
//	@Summary		Check server health
//	@Description	Returns a JSON response indicating the service status and current timestamp.
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/ [get]
func healthcheck(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "not found",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := map[string]string{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(resp)
}

// Creates an API handler, expected to be used once during the initialization of the application
func Api(idempotentRequestHandler *IdempotentRequestHandler, idempotencyLayer internal.IdempotencyLayer) *http.ServeMux {
	v1Mux := v1.Routes(idempotencyLayer)

	api := http.NewServeMux()
	api.Handle("/docs/", httpSwagger.WrapHandler)

	baseHandler := http.StripPrefix("/v1", v1Mux)
	idempotentHandler := idempotentRequestHandler.HandleRequest(baseHandler)
	loggedHandler := CanonicalLogger(idempotentHandler)
	finalHandler := RequestMetadataMiddleware(loggedHandler)

	api.HandleFunc("/v1/", finalHandler)
	api.HandleFunc("/", healthcheck)

	return api
}
