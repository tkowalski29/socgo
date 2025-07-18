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
	oauthService := oauth.NewService(container.GetDBManager(), container.GetConfig())
	oauthHandler := oauth.NewHandler(oauthService)

	// Post handler (API)
	postHandler := handlers.NewPostHandler(container.GetDBManager(), container.GetProviderService())

	// Web handler (UI)
	webHandler := handlers.NewWebHandler(container.GetDBManager(), container.GetProviderService())

	// API token handler
	apiTokenHandler := handlers.NewAPITokenHandler(container.GetDBManager())

	// Auth middleware
	authMiddleware := middleware.NewAuthMiddleware(container.GetDBManager())

	// Web routes (UI pages)
	r.HandleFunc("/", webHandler.HomePage).Methods("GET")
	r.HandleFunc("/dashboard", webHandler.DashboardPage).Methods("GET")
	r.HandleFunc("/providers", webHandler.ProvidersPage).Methods("GET")
	r.HandleFunc("/posts", webHandler.PostsPage).Methods("GET")
	r.HandleFunc("/calendar", webHandler.CalendarPage).Methods("GET")
	r.HandleFunc("/health", handlers.HealthHandler)

	// Web form handlers
	r.HandleFunc("/posts", webHandler.HandlePost).Methods("POST")

	// HTMX/AJAX endpoints for web UI
	r.HandleFunc("/posts/history", postHandler.HandleHistory).Methods("GET")
	r.HandleFunc("/posts/calendar", postHandler.HandleCalendar).Methods("GET")
	r.HandleFunc("/posts/calendar-page", postHandler.HandleCalendarPage).Methods("GET")

	// Stats endpoints for dashboard
	r.HandleFunc("/api/stats/providers", webHandler.HandleProvidersCount).Methods("GET")
	r.HandleFunc("/api/stats/published", webHandler.HandlePublishedCount).Methods("GET")
	r.HandleFunc("/api/stats/scheduled", webHandler.HandleScheduledCount).Methods("GET")
	r.HandleFunc("/api/stats/monthly", webHandler.HandleMonthlyCount).Methods("GET")
	r.HandleFunc("/api/providers/options", webHandler.HandleProvidersOptions).Methods("GET")

	// OAuth routes
	r.HandleFunc("/connect/{provider}", oauthHandler.HandleConnect).Methods("GET")
	r.HandleFunc("/oauth/callback/{provider}", oauthHandler.HandleCallback).Methods("GET")
	r.HandleFunc("/api/providers/available", oauthHandler.HandleAvailableProviders).Methods("GET")
	r.HandleFunc("/api/providers", oauthHandler.HandleProviders).Methods("GET")
	r.HandleFunc("/api/providers/{id}", oauthHandler.HandleDisconnect).Methods("DELETE")

	// API token generation endpoint (public)
	r.HandleFunc("/api-tokens", apiTokenHandler.HandleCreateToken).Methods("POST")

	// Protected API routes with auth middleware
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.Use(authMiddleware.APIAuthMiddleware)

	// JSON API endpoints (for external integrations)
	apiRouter.HandleFunc("/posts", postHandler.HandlePost).Methods("POST")

	return r
}
