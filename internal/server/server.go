package server

import (
	"net/http"

	"github.com/tkowalski/socgo/internal/di"
	"github.com/tkowalski/socgo/internal/handlers"
)

func New(container *di.Container) http.Handler {
	mux := http.NewServeMux()
	
	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	
	// Routes
	mux.HandleFunc("/", handlers.HomeHandler)
	mux.HandleFunc("/health", handlers.HealthHandler)
	
	return mux
}