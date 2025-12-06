package v1

import (
	"fmt"
	"net/http"
)

func ping(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "pong!\n")
}

func Routes() *http.ServeMux {
	v1 := http.NewServeMux()

	v1.HandleFunc("/ping", ping)

	return v1
}
