package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tkowalski/socgo/internal/handlers"
	"github.com/tkowalski/socgo/internal/service/dependency"
	"github.com/tkowalski/socgo/internal/service/oauth"
	"github.com/tkowalski/socgo/internal/service/server/internal"
	"github.com/tkowalski/socgo/internal/service/server/middleware"
)

func New(container *dependency.Container) http.Handler {
	r := mux.NewRouter()

	// Add compression middleware
	r.Use(internal.GzipMiddleware)

	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	// OAuth service and handler
	oauthService := oauth.NewService(container.GetDBManager(), container.GetConfig())
	oauthHandler := handlers.NewOAuthHandler(oauthService)

	// Post handler (API)
	postHandler := handlers.NewPostHandler(container.GetDBManager(), container.GetProviderService())

	// Calendar handler
	calendarHandler := handlers.NewCalendarHandler(container.GetDBManager(), container.GetProviderService())

	// Web handler (UI)
	settingHandler := handlers.NewSettingHandler(container.GetDBManager(), container.GetProviderService(), container.GetNotificationService(), oauthService)

	// API token handler
	apiTokenHandler := handlers.NewAPITokenHandler(container.GetDBManager())

	// Notification handler
	notificationHandler := handlers.NewNotificationHandler(container.GetNotificationService())

	// Auth middleware
	authMiddleware := middleware.NewAuthMiddleware(container.GetDBManager())

	// Web routes (UI pages)
	r.HandleFunc("/", calendarHandler.CalendarPage).Methods("GET")
	r.HandleFunc("/settings", settingHandler.SettingsPage).Methods("GET")
	r.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("pong")); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	})

	// Web form handlers
	r.HandleFunc("/posts", postHandler.HandlePost).Methods("POST")

	// HTMX/AJAX endpoints for web UI
	r.HandleFunc("/posts/history", postHandler.HandleHistory).Methods("GET")
	r.HandleFunc("/posts/calendar", calendarHandler.HandleCalendar).Methods("GET")
	r.HandleFunc("/api/calendar/week", calendarHandler.HandleWeekView).Methods("GET")
	r.HandleFunc("/api/posts/{id}", postHandler.HandlePostDetails).Methods("GET")
	r.HandleFunc("/api/posts/{id}", postHandler.HandleDeletePost).Methods("DELETE")

	// Stats endpoints for dashboard
	r.HandleFunc("/api/stats/providers", postHandler.HandleProvidersCount).Methods("GET")
	r.HandleFunc("/api/stats/published", postHandler.HandlePublishedCount).Methods("GET")
	r.HandleFunc("/api/stats/scheduled", postHandler.HandleScheduledCount).Methods("GET")
	r.HandleFunc("/api/stats/monthly", postHandler.HandleMonthlyCount).Methods("GET")
	r.HandleFunc("/api/providers/options", postHandler.HandleProvidersOptions).Methods("GET")
	r.HandleFunc("/api/providers/settings", postHandler.HandleProviderSettings).Methods("GET")
	r.HandleFunc("/api/providers/tabs", postHandler.HandleProviderTabs).Methods("GET")
	r.HandleFunc("/api/posts/success", postHandler.HandlePostSuccess).Methods("GET")
	r.HandleFunc("/api/file-preview", postHandler.HandleFilePreview).Methods("POST")

	// Notification routes
	r.HandleFunc("/api/notifications/stats", notificationHandler.HandleGetNotificationStats).Methods("GET")
	r.HandleFunc("/api/notifications/groups", notificationHandler.HandleGetNotificationGroups).Methods("GET")
	r.HandleFunc("/api/notifications/groups/{groupID}", notificationHandler.HandleGetNotificationsByGroup).Methods("GET")
	r.HandleFunc("/api/notifications/groups/{groupID}/read", notificationHandler.HandleMarkGroupAsRead).Methods("POST")

	// Notification UI routes
	r.HandleFunc("/api/notifications/bell", notificationHandler.HandleNotificationBell).Methods("GET")
	r.HandleFunc("/api/notifications/groups-ui", notificationHandler.HandleNotificationGroups).Methods("GET")
	r.HandleFunc("/api/notifications/details/{groupID}", notificationHandler.HandleNotificationDetails).Methods("GET")

	// OAuth routes
	r.HandleFunc("/connect/{provider}", oauthHandler.HandleConnect).Methods("GET")
	r.HandleFunc("/oauth/callback/{provider}", oauthHandler.HandleCallback).Methods("GET")
	r.HandleFunc("/api/providers/available", settingHandler.HandleAvailableProviders).Methods("GET")
	r.HandleFunc("/api/providers", settingHandler.HandleProviders).Methods("GET")
	r.HandleFunc("/api/providers/{id}", settingHandler.HandleDisconnect).Methods("DELETE")

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
