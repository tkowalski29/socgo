package server

import (
	"net/http"

	"github.com/tkowalski/socgo/internal/handlers"
)

func New() http.Handler {
	mux := http.NewServeMux()
	
	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	
	// Routes
	mux.HandleFunc("/", handlers.HomeHandler)
	mux.HandleFunc("/health", handlers.HealthHandler)
	
	return mux
}