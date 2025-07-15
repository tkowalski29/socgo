package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tkowalski/socgo/internal/di"
	"github.com/tkowalski/socgo/internal/handlers"
	"github.com/tkowalski/socgo/internal/oauth"
)

func New(container *di.Container) http.Handler {
	r := mux.NewRouter()

	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// OAuth service and handler
	oauthService := oauth.NewService(container.GetDBManager())
	oauthHandler := oauth.NewHandler(oauthService)

	// Post handler
	postHandler := handlers.NewPostHandler(container.GetDBManager(), container.GetProviderService())

	// Routes
	r.HandleFunc("/", handlers.HomeHandler)
	r.HandleFunc("/health", handlers.HealthHandler)

	// Post routes
	r.HandleFunc("/posts", postHandler.HandlePost).Methods("POST")

	// OAuth routes
	r.HandleFunc("/connect/{provider}", oauthHandler.HandleConnect).Methods("GET")
	r.HandleFunc("/oauth/callback/{provider}", oauthHandler.HandleCallback).Methods("GET")
	r.HandleFunc("/providers", oauthHandler.HandleProviders).Methods("GET")
	r.HandleFunc("/providers/{id}", oauthHandler.HandleDisconnect).Methods("DELETE")

	return r
}
