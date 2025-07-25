package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tkowalski/socgo/internal/di"
	"github.com/tkowalski/socgo/internal/handlers"
	"github.com/tkowalski/socgo/internal/service/oauth"
	"github.com/tkowalski/socgo/internal/service/server/middleware"
)

func New(container *di.Container) http.Handler {
	r := mux.NewRouter()

	// Add compression middleware
	r.Use(handlers.GzipMiddleware)

	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	// OAuth service and handler
	oauthService := oauth.NewService(container.GetDBManager(), container.GetConfig())
	oauthHandler := oauth.NewHandler(oauthService)

	// Post handler (API)
	postHandler := handlers.NewPostHandler(container.GetDBManager(), container.GetProviderService())

	// Web handler (UI)
	webHandler := handlers.NewWebHandler(container.GetDBManager(), container.GetProviderService(), container.GetNotificationService())

	// API token handler
	apiTokenHandler := handlers.NewAPITokenHandler(container.GetDBManager())

	// Notification handler
	notificationHandler := handlers.NewNotificationHandler(container.GetNotificationService())

	// Auth middleware
	authMiddleware := middleware.NewAuthMiddleware(container.GetDBManager())

	// Web routes (UI pages)
	r.HandleFunc("/calendar", webHandler.CalendarPage).Methods("GET")
	r.HandleFunc("/settings", webHandler.SettingsPage).Methods("GET")
	r.HandleFunc("/health", handlers.HealthHandler)

	// Web form handlers
	r.HandleFunc("/posts", webHandler.HandlePost).Methods("POST")

	// HTMX/AJAX endpoints for web UI
	r.HandleFunc("/posts/history", postHandler.HandleHistory).Methods("GET")
	r.HandleFunc("/posts/calendar", postHandler.HandleCalendar).Methods("GET")
	r.HandleFunc("/posts/calendar-page", postHandler.HandleCalendarPage).Methods("GET")
	r.HandleFunc("/api/calendar/week", postHandler.HandleWeekView).Methods("GET")
	r.HandleFunc("/api/posts/{id}", postHandler.HandlePostDetails).Methods("GET")
	r.HandleFunc("/api/posts/{id}", postHandler.HandleDeletePost).Methods("DELETE")

	// Stats endpoints for dashboard
	r.HandleFunc("/api/stats/providers", webHandler.HandleProvidersCount).Methods("GET")
	r.HandleFunc("/api/stats/published", webHandler.HandlePublishedCount).Methods("GET")
	r.HandleFunc("/api/stats/scheduled", webHandler.HandleScheduledCount).Methods("GET")
	r.HandleFunc("/api/stats/monthly", webHandler.HandleMonthlyCount).Methods("GET")
	r.HandleFunc("/api/providers/options", webHandler.HandleProvidersOptions).Methods("GET")
	r.HandleFunc("/api/providers/settings", webHandler.HandleProviderSettings).Methods("GET")
	r.HandleFunc("/api/providers/tabs", webHandler.HandleProviderTabs).Methods("GET")
	r.HandleFunc("/api/posts/success", webHandler.HandlePostSuccess).Methods("GET")
	r.HandleFunc("/api/error-message", webHandler.HandleErrorMessage).Methods("GET")
	r.HandleFunc("/api/cache/clear", webHandler.HandleCacheClear).Methods("POST")
	r.HandleFunc("/api/cache/stats", webHandler.HandleCacheStats).Methods("GET")
	r.HandleFunc("/api/file-preview", webHandler.HandleFilePreview).Methods("POST")

	// Notification routes
	r.HandleFunc("/api/notifications/stats", notificationHandler.HandleGetNotificationStats).Methods("GET")
	r.HandleFunc("/api/notifications/groups", notificationHandler.HandleGetNotificationGroups).Methods("GET")
	r.HandleFunc("/api/notifications/groups/{groupID}", notificationHandler.HandleGetNotificationsByGroup).Methods("GET")
	r.HandleFunc("/api/notifications/groups/{groupID}/read", notificationHandler.HandleMarkGroupAsRead).Methods("POST")

	// Notification UI routes
	r.HandleFunc("/api/notifications/bell", webHandler.HandleNotificationBell).Methods("GET")
	r.HandleFunc("/api/notifications/groups-ui", webHandler.HandleNotificationGroups).Methods("GET")
	r.HandleFunc("/api/notifications/details/{groupID}", webHandler.HandleNotificationDetails).Methods("GET")

	// OAuth routes
	r.HandleFunc("/connect/{provider}", oauthHandler.HandleConnect).Methods("GET")
	r.HandleFunc("/oauth/callback/{provider}", oauthHandler.HandleCallback).Methods("GET")
	r.HandleFunc("/api/providers/available", oauthHandler.HandleAvailableProviders).Methods("GET")
	r.HandleFunc("/api/providers", oauthHandler.HandleProviders).Methods("GET")
	r.HandleFunc("/api/providers/{id}", oauthHandler.HandleDisconnect).Methods("DELETE")

	// API token generation endpoint (public)
	r.HandleFunc("/api-tokens", apiTokenHandler.HandleCreateToken).Methods("POST")
	r.HandleFunc("/api/tokens/{id}", apiTokenHandler.HandleDeleteToken).Methods("DELETE")

	// Protected API routes with auth middleware
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.Use(authMiddleware.APIAuthMiddleware)

	// JSON API endpoints (for external integrations)
	apiRouter.HandleFunc("/posts", postHandler.HandlePost).Methods("POST")

	return r
}
