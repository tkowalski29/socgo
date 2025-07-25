package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	data_database "github.com/tkowalski/socgo/internal/data/database"
	"github.com/tkowalski/socgo/internal/handlers/internal"
	"github.com/tkowalski/socgo/internal/service/database"
	"github.com/tkowalski/socgo/internal/service/notifications"
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
}

// NewSettingHandler creates a new SettingHandler instance
func NewSettingHandler(dbManager *database.Manager, providerService *post.ProviderService, notificationService *notifications.Service) *SettingHandler {
	return &SettingHandler{
		dbManager:           dbManager,
		providerService:     providerService,
		notificationService: notificationService,
		cache:               internal.NewComponentCache(),
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
	var userTokens []component.UserToken
	if activeSection == "user" {
		userID := internal.GetUserID(r)
		db, err := h.dbManager.GetDB(userID)
		if err == nil {
			var apiTokens []data_database.APIToken
			if err := db.Where("user_id = ?", userID).Find(&apiTokens).Error; err == nil {
				for _, token := range apiTokens {
					userToken := component.UserToken{
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
