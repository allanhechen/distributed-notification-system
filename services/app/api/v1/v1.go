// The v1 API handlers for the application
// Serves as the jumping-off point for other modules (ex. users, devices, groups)
package v1

import (
	"fmt"
	"net/http"
)

// Connectivity Check
//
//	@Summary		Test server connectivity
//	@Description	Endpoint to quickly verify that the server is reachable
//	@Tags			health
//	@Produce		plain
//	@Success		200	{string}	string	"pong!"
// ping is an HTTP handler for a simple health check that writes "pong!\n" to the response.
func ping(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "pong!\n")
}

// Routes creates and returns an HTTP ServeMux configured with v1 API routes.
// It registers the /ping health-check endpoint and is intended to be called
// once during application initialization to mount v1 handlers.
func Routes() *http.ServeMux {
	v1 := http.NewServeMux()

	v1.HandleFunc("/ping", ping)

	return v1
}