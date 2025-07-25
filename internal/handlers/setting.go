package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	data_database "github.com/tkowalski/socgo/internal/data/database"
	"github.com/tkowalski/socgo/internal/handlers/internal"
	"github.com/tkowalski/socgo/internal/service/database"
	"github.com/tkowalski/socgo/internal/service/notifications"
	"github.com/tkowalski/socgo/internal/service/oauth"
	"github.com/tkowalski/socgo/internal/service/post"
	"github.com/tkowalski/socgo/web/templates"
	"github.com/tkowalski/socgo/web/templates/component"
)

// SettingHandler handles web page requests
type SettingHandler struct {
	dbManager           *database.Manager
	providerService     *post.ProviderService
	notificationService *notifications.Service
	cache               *internal.ComponentCache
	oauthService        *oauth.Service
}

// NewSettingHandler creates a new SettingHandler instance
func NewSettingHandler(dbManager *database.Manager, providerService *post.ProviderService, notificationService *notifications.Service, oauthService *oauth.Service) *SettingHandler {
	return &SettingHandler{
		dbManager:           dbManager,
		providerService:     providerService,
		notificationService: notificationService,
		cache:               internal.NewComponentCache(),
		oauthService:        oauthService,
	}
}

// SettingsPage handles the settings page
func (h *SettingHandler) SettingsPage(w http.ResponseWriter, r *http.Request) {
	// Get flash message from query parameters
	flashMessage := ""
	flashType := "info"
	if flashMsg := r.URL.Query().Get("flash"); flashMsg != "" {
		if decoded, err := url.QueryUnescape(flashMsg); err == nil {
			flashMessage = decoded
		} else {
			flashMessage = flashMsg
		}
		flashType = r.URL.Query().Get("flash_type")
		if flashType == "" {
			flashType = "info"
		}
	}

	// Get active section from query parameters (default to "user")
	activeSection := r.URL.Query().Get("section")
	if activeSection == "" {
		activeSection = "user"
	}

	// Get user tokens for user section
	var userTokens []templates.UserToken
	if activeSection == "user" {
		userID := internal.GetUserID(r)
		db, err := h.dbManager.GetDB(userID)
		if err == nil {
			var apiTokens []data_database.APIToken
			if err := db.Where("user_id = ?", userID).Find(&apiTokens).Error; err == nil {
				for _, token := range apiTokens {
					userToken := templates.UserToken{
						ID:        fmt.Sprintf("%d", token.ID),
						Name:      fmt.Sprintf("Token #%d", token.ID),
						CreatedAt: token.CreatedAt.Format("2006-01-02 15:04:05"),
						LastUsed:  "",
					}
					if token.LastUsed != nil {
						userToken.LastUsed = token.LastUsed.Format("2006-01-02 15:04:05")
					}
					userTokens = append(userTokens, userToken)
				}
			}
		}
	}

	// Create settings data
	settingsData := templates.SettingsData{
		ActiveSection: activeSection,
		UserTokens:    userTokens,
	}

	// Create layout data
	layoutData := templates.LayoutData{
		Title:        "Ustawienia",
		CurrentPage:  "settings",
		FlashMessage: flashMessage,
		FlashType:    flashType,
		Content:      templates.SettingsContent(settingsData),
	}

	// Render the layout
	w.Header().Set("Content-Type", "text/html")
	layoutComponent := templates.Layout(layoutData)
	if err := layoutComponent.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering settings page: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// HandleAvailableProviders handles the available providers page
func (h *SettingHandler) HandleAvailableProviders(w http.ResponseWriter, r *http.Request) {
	userID := internal.GetUserID(r)
	if userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	availableProviders := h.oauthService.GetAvailableProviders()

	w.Header().Set("Content-Type", "text/html")

	if err := component.ProviderAvailable(availableProviders).Render(r.Context(), w); err != nil {
		log.Printf("Error rendering available providers: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *SettingHandler) HandleProviders(w http.ResponseWriter, r *http.Request) {
	userID := internal.GetUserID(r)
	if userID == "" {
		// For API requests, return JSON error
		if strings.Contains(r.Header.Get("Accept"), "application/json") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			if _, err := w.Write([]byte(`{"error": "User not authenticated"}`)); err != nil {
				log.Printf("Error writing response: %v", err)
			}
			return
		}
		// For HTML requests, redirect with error
		errorMsg := url.QueryEscape("User not authenticated")
		http.Redirect(w, r, "/?flash="+errorMsg+"&flash_type=error", http.StatusTemporaryRedirect)
		return
	}

	providers, err := h.oauthService.GetProviders(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get providers: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if request wants JSON response
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "application/json") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"message": "Connected providers"}`)); err != nil {
			log.Printf("Error writing response: %v", err)
		}
		return
	}

	w.Header().Set("Content-Type", "text/html")

	if err := component.ProviderConnected(providers).Render(r.Context(), w); err != nil {
		log.Printf("Error rendering connected providers: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *SettingHandler) HandleDisconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	providerIDStr := vars["id"]
	providerID, err := strconv.ParseUint(providerIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid provider ID", http.StatusBadRequest)
		return
	}

	userID := internal.GetUserID(r)
	if userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	err = h.oauthService.DisconnectProvider(userID, uint(providerID))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to disconnect provider: %v", err), http.StatusInternalServerError)
		return
	}

	// Return empty response and trigger list refresh for HTMX
	w.Header().Set("HX-Trigger", "refresh-providers")
	w.WriteHeader(http.StatusOK)
}
