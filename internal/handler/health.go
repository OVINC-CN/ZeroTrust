package handler

import (
	"net/http"

	"github.com/ovinc/zerotrust/internal/store"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	// store ping
	if err := store.Ping(r.Context()); err != nil {
		http.Error(w, "store unreachable", http.StatusServiceUnavailable)
		return
	}

	// respond with 200 OK
	w.WriteHeader(http.StatusOK)
}
