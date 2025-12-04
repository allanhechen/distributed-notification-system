package main

import (
	"fmt"
	"net/http"
)

func ping(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "pong!\n")
}

func main() {
	http.HandleFunc("/ping", ping)
	http.ListenAndServe(":8080", nil)
}
