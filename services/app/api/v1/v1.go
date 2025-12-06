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
//	@Router			/v1/ping [get]
func ping(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "pong!\n")
}

// Creates a v1 Routes handler, expected to be used once during the initialization of the application
func Routes() *http.ServeMux {
	v1 := http.NewServeMux()

	v1.HandleFunc("/ping", ping)

	return v1
}
