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
func Api() *http.ServeMux {
	v1Mux := v1.Routes()

	api := http.NewServeMux()
	api.Handle("/docs/", httpSwagger.WrapHandler)

	api.HandleFunc("/v1/", RequestMetadataMiddleware(CanonicalLogger(http.StripPrefix("/v1", v1Mux))))
	api.HandleFunc("/", healthcheck)

	return api
}
