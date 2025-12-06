package api

import (
	"encoding/json"
	"net/http"
	"time"

	v1 "github.com/allanhechen/distributed-notification-system/services/app/api/v1"
)

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

func Api() *http.ServeMux {
	v1 := v1.Routes()

	api := http.NewServeMux()
	api.Handle("/v1/", http.StripPrefix("/v1", v1))
	api.HandleFunc("/", healthcheck)

	return api
}
