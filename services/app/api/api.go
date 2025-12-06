// Serves as the main entrypoint into the API handlers of the application
package api

import (
	"encoding/json"
	"net/http"
	"time"

	v1 "github.com/allanhechen/distributed-notification-system/services/app/api/v1"
	_ "github.com/allanhechen/distributed-notification-system/services/app/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Healthcheck endpoint
//
//	@Summary		Check server health
//	@Description	Returns a JSON response indicating the service status and current timestamp.
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	map[string]string
// healthcheck responds to requests to the root path with a JSON status and current timestamp.
// If the request URL path is not exactly "/", it responds with HTTP 404 and a JSON error object.
func healthcheck(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
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

// Api constructs and returns an *http.ServeMux configured with the application's HTTP routes.
// It mounts Swagger documentation at /docs/, serves version 1 routes under /v1/, and registers
// a root healthcheck at /. Intended to be called once during application initialization.
func Api() *http.ServeMux {
	v1 := v1.Routes()

	api := http.NewServeMux()
	api.Handle("/docs/", httpSwagger.WrapHandler)

	api.Handle("/v1/", http.StripPrefix("/v1", v1))
	api.HandleFunc("/", healthcheck)

	return api
}