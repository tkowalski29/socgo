package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tkowalski/socgo/internal/di"
	"github.com/tkowalski/socgo/internal/handlers"
	"github.com/tkowalski/socgo/internal/middleware"
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

	// API token handler
	apiTokenHandler := handlers.NewAPITokenHandler(container.GetDBManager())

	// Auth middleware
	authMiddleware := middleware.NewAuthMiddleware(container.GetDBManager())

	// Routes
	r.HandleFunc("/", handlers.HomeHandler)
	r.HandleFunc("/health", handlers.HealthHandler)

	// Post routes
	r.HandleFunc("/posts", postHandler.HandlePost).Methods("POST")
	r.HandleFunc("/posts/history", postHandler.HandleHistory).Methods("GET")
	r.HandleFunc("/posts/calendar", postHandler.HandleCalendar).Methods("GET")
	r.HandleFunc("/posts/calendar-page", postHandler.HandleCalendarPage).Methods("GET")

	// OAuth routes
	r.HandleFunc("/connect/{provider}", oauthHandler.HandleConnect).Methods("GET")
	r.HandleFunc("/oauth/callback/{provider}", oauthHandler.HandleCallback).Methods("GET")
	r.HandleFunc("/providers", oauthHandler.HandleProviders).Methods("GET")
	r.HandleFunc("/providers/{id}", oauthHandler.HandleDisconnect).Methods("DELETE")

	// API token generation endpoint (public)
	r.HandleFunc("/api-tokens", apiTokenHandler.HandleCreateToken).Methods("POST")

	// Protected API routes with auth middleware
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.Use(authMiddleware.APIAuthMiddleware)
	
	// Move posts endpoint to protected API routes
	apiRouter.HandleFunc("/posts", postHandler.HandlePost).Methods("POST")

	return r
}
