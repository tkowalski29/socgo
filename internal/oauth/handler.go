package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/tkowalski/socgo/internal/config"
	"github.com/tkowalski/socgo/internal/database"
)

type Handler struct {
	oauthService *Service
}

func NewHandler(oauthService *Service) *Handler {
	return &Handler{
		oauthService: oauthService,
	}
}

func (h *Handler) HandleConnect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider"]
	providerName := r.URL.Query().Get("name")

	if providerName == "" {
		// Redirect with error message
		errorMsg := url.QueryEscape("Provider name is required")
		http.Redirect(w, r, "/providers?flash="+errorMsg+"&flash_type=error", http.StatusTemporaryRedirect)
		return
	}

	providerType := ProviderType(strings.ToLower(provider))

	if _, exists := SupportedProviders[providerType]; !exists {
		// Redirect with error message
		errorMsg := url.QueryEscape(fmt.Sprintf("Unsupported provider: %s", provider))
		http.Redirect(w, r, "/providers?flash="+errorMsg+"&flash_type=error", http.StatusTemporaryRedirect)
		return
	}

	userID := h.getUserID(r)
	if userID == "" {
		// Redirect with error message
		errorMsg := url.QueryEscape("User not authenticated")
		http.Redirect(w, r, "/providers?flash="+errorMsg+"&flash_type=error", http.StatusTemporaryRedirect)
		return
	}

	connectURL, err := h.oauthService.GetConnectURL(userID, providerType, providerName)
	if err != nil {
		// Redirect with error message
		errorMsg := url.QueryEscape(fmt.Sprintf("Failed to generate connect URL: %v", err))
		http.Redirect(w, r, "/providers?flash="+errorMsg+"&flash_type=error", http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, connectURL, http.StatusTemporaryRedirect)
}

func (h *Handler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider"]

	providerType := ProviderType(strings.ToLower(provider))

	if _, exists := SupportedProviders[providerType]; !exists {
		// Redirect with error message
		errorMsg := url.QueryEscape(fmt.Sprintf("Unsupported provider: %s", provider))
		http.Redirect(w, r, "/providers?flash="+errorMsg+"&flash_type=error", http.StatusTemporaryRedirect)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		// Redirect with error message
		errorMsg := url.QueryEscape("Authorization code not provided")
		http.Redirect(w, r, "/providers?flash="+errorMsg+"&flash_type=error", http.StatusTemporaryRedirect)
		return
	}

	state := r.URL.Query().Get("state")
	if state == "" {
		// Redirect with error message
		errorMsg := url.QueryEscape("State parameter missing")
		http.Redirect(w, r, "/providers?flash="+errorMsg+"&flash_type=error", http.StatusTemporaryRedirect)
		return
	}

	// Parse state to get userID and providerName
	stateParts := strings.Split(state, ":")
	if len(stateParts) != 2 {
		// Redirect with error message
		errorMsg := url.QueryEscape("Invalid state parameter")
		http.Redirect(w, r, "/providers?flash="+errorMsg+"&flash_type=error", http.StatusTemporaryRedirect)
		return
	}

	userID := stateParts[0]
	providerName := stateParts[1]

	err := h.oauthService.HandleCallback(userID, providerType, code, providerName)
	if err != nil {
		// Redirect with error message
		errorMsg := url.QueryEscape(fmt.Sprintf("Failed to connect provider: %v", err))
		http.Redirect(w, r, "/providers?flash="+errorMsg+"&flash_type=error", http.StatusTemporaryRedirect)
		return
	}

	// Redirect with success message
	successMsg := url.QueryEscape(fmt.Sprintf("Successfully connected to %s (%s)", providerName, providerType))
	http.Redirect(w, r, "/providers?flash="+successMsg+"&flash_type=success", http.StatusTemporaryRedirect)
}

func (h *Handler) HandleProviders(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	if userID == "" {
		// For API requests, return JSON error
		if strings.Contains(r.Header.Get("Accept"), "application/json") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "User not authenticated"}`))
			return
		}
		// For HTML requests, redirect with error
		errorMsg := url.QueryEscape("User not authenticated")
		http.Redirect(w, r, "/providers?flash="+errorMsg+"&flash_type=error", http.StatusTemporaryRedirect)
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
		h.handleProvidersJSON(w, providers)
		return
	}

	h.handleProvidersHTML(w, providers)
}

func (h *Handler) HandleAvailableProviders(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	if userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	availableProviders := h.oauthService.GetAvailableProviders()
	h.handleAvailableProvidersHTML(w, availableProviders)
}

func (h *Handler) handleProvidersJSON(w http.ResponseWriter, providers []database.Provider) {
	type ProviderResponse struct {
		DisplayName  string `json:"display_name"`
		ProviderType string `json:"provider_type"`
		ConnectedAt  string `json:"connected_at"`
		Status       string `json:"status"`
	}

	response := make([]ProviderResponse, len(providers))
	for i, provider := range providers {
		// Parse config to get user info for display name
		var config ProviderConfig
		displayName := provider.Name
		if provider.Config != "" {
			if err := json.Unmarshal([]byte(provider.Config), &config); err == nil {
				if config.UserInfo != nil && config.UserInfo.Name != "" {
					displayName = config.UserInfo.Name
				}
			}
		}

		status := "active"
		if !provider.IsActive {
			status = "inactive"
		}

		response[i] = ProviderResponse{
			DisplayName:  displayName,
			ProviderType: provider.Type,
			ConnectedAt:  provider.CreatedAt.Format("2006-01-02T15:04:05Z"),
			Status:       status,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) handleProvidersHTML(w http.ResponseWriter, providers []database.Provider) {
	w.Header().Set("Content-Type", "text/html")

	if len(providers) == 0 {
		html := `<div class="text-center py-8">
			<div class="w-16 h-16 bg-gray-100 rounded-full mx-auto mb-4 flex items-center justify-center">
				<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
				</svg>
			</div>
			<h3 class="text-lg font-medium text-gray-900 mb-2">No providers connected</h3>
			<p class="text-gray-500">Connect your first social media account to get started.</p>
		</div>`
		if _, err := w.Write([]byte(html)); err != nil {
			// Log error but don't fail the operation since response headers are already set
			_ = err // explicitly ignore error
		}
		return
	}

	html := `<div class="space-y-4">`
	for _, provider := range providers {
		// Parse config to get user info for display name
		var config ProviderConfig
		displayName := provider.Name
		if provider.Config != "" {
			if err := json.Unmarshal([]byte(provider.Config), &config); err == nil {
				if config.UserInfo != nil && config.UserInfo.Name != "" {
					displayName = config.UserInfo.Name
				}
			}
		}

		status := "Active"
		statusClass := "bg-green-100 text-green-800"
		if !provider.IsActive {
			status = "Inactive"
			statusClass = "bg-red-100 text-red-800"
		}

		// Get provider icon based on type
		iconClass := "bg-gray-500"
		switch provider.Type {
		case "tiktok":
			iconClass = "bg-black"
		case "instagram":
			iconClass = "bg-gradient-to-r from-purple-500 to-pink-500"
		case "facebook":
			iconClass = "bg-blue-600"
		}

		html += fmt.Sprintf(`
			<div class="border rounded-lg p-4 flex items-center justify-between hover:shadow-md transition-shadow">
				<div class="flex items-center space-x-4">
					<div class="w-10 h-10 %s rounded-full flex items-center justify-center">
						<span class="text-white font-semibold text-sm capitalize">%s</span>
					</div>
					<div>
						<h3 class="font-semibold text-gray-900">%s</h3>
						<p class="text-sm text-gray-600">Connected on %s</p>
					</div>
				</div>
				<div class="flex items-center space-x-3">
					<span class="%s px-2 py-1 rounded-full text-xs font-medium">%s</span>
					<button hx-delete="/api/providers/%d" hx-target="closest div" hx-swap="outerHTML" hx-confirm="Are you sure you want to disconnect this provider?" class="bg-red-500 hover:bg-red-700 text-white font-medium py-1 px-3 rounded text-sm transition-colors">
						Disconnect
					</button>
				</div>
			</div>`, iconClass, provider.Type[:3], displayName, provider.CreatedAt.Format("January 2, 2006"), statusClass, status, provider.ID)
	}
	html += `</div>`

	if _, err := w.Write([]byte(html)); err != nil {
		// Log error but don't fail the operation since response headers are already set
		_ = err // explicitly ignore error
	}
}

func (h *Handler) handleAvailableProvidersHTML(w http.ResponseWriter, availableProviders map[string][]config.ProviderInstance) {
	w.Header().Set("Content-Type", "text/html")

	html := `<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">`

	// TikTok providers
	for _, provider := range availableProviders["tiktok"] {
		html += fmt.Sprintf(`
			<div class="border rounded-lg p-4 text-center hover:shadow-md transition-shadow">
				<div class="w-12 h-12 bg-black rounded-full mx-auto mb-3 flex items-center justify-center">
					<svg class="w-6 h-6 text-white" fill="currentColor" viewBox="0 0 24 24">
						<path d="M12.525.02c1.31-.02 2.61-.01 3.91-.02.08 1.53.63 3.09 1.75 4.17 1.12 1.11 2.7 1.62 4.24 1.79v4.03c-1.44-.05-2.89-.35-4.2-.97-.57-.26-1.1-.59-1.62-.93-.01 2.92.01 5.84-.02 8.75-.08 1.4-.54 2.79-1.35 3.94-1.31 1.92-3.58 3.17-5.91 3.21-1.43.08-2.86-.31-4.08-1.03-2.02-1.19-3.44-3.37-3.65-5.71-.02-.5-.03-1-.01-1.49.18-1.9 1.12-3.72 2.58-4.96 1.66-1.44 3.98-2.13 6.15-1.72.02 1.48-.04 2.96-.04 4.44-.99-.32-2.15-.23-3.02.37-.63.41-1.11 1.04-1.36 1.75-.21.51-.15 1.07-.14 1.61.24 1.64 1.82 3.02 3.5 2.87 1.12-.01 2.19-.66 2.77-1.61.19-.33.4-.67.41-1.06.1-1.79.06-3.57.07-5.36.01-4.03-.01-8.05.02-12.07z"/>
					</svg>
				</div>
				<h3 class="font-semibold mb-2">%s</h3>
				<p class="text-sm text-gray-600 mb-3">%s</p>
				<a href="/connect/tiktok?name=%s" class="inline-block bg-black text-white px-4 py-2 rounded-lg hover:bg-gray-800 transition-colors text-sm">
					Connect TikTok
				</a>
			</div>`, provider.Name, provider.Description, provider.Name)
	}

	// Instagram providers
	for _, provider := range availableProviders["instagram"] {
		html += fmt.Sprintf(`
			<div class="border rounded-lg p-4 text-center hover:shadow-md transition-shadow">
				<div class="w-12 h-12 bg-gradient-to-r from-purple-500 to-pink-500 rounded-full mx-auto mb-3 flex items-center justify-center">
					<svg class="w-6 h-6 text-white" fill="currentColor" viewBox="0 0 24 24">
						<path d="M12 2.163c3.204 0 3.584.012 4.85.07 3.252.148 4.771 1.691 4.919 4.919.058 1.265.069 1.645.069 4.849 0 3.205-.012 3.584-.069 4.849-.149 3.225-1.664 4.771-4.919 4.919-1.266.058-1.644.07-4.85.07-3.204 0-3.584-.012-4.849-.07-3.26-.149-4.771-1.699-4.919-4.92-.058-1.265-.07-1.644-.07-4.849 0-3.204.013-3.583.07-4.849.149-3.227 1.664-4.771 4.919-4.919 1.266-.057 1.645-.069 4.849-.069zm0-2.163c-3.259 0-3.667.014-4.947.072-4.358.2-6.78 2.618-6.98 6.98-.059 1.281-.073 1.689-.073 4.948 0 3.259.014 3.668.072 4.948.2 4.358 2.618 6.78 6.98 6.98 1.281.058 1.689.072 4.948.072 3.259 0 3.668-.014 4.948-.072 4.354-.2 6.782-2.618 6.979-6.98.059-1.28.073-1.689.073-4.948 0-3.259-.014-3.667-.072-4.947-.196-4.354-2.617-6.78-6.979-6.98-1.281-.059-1.69-.073-4.949-.073zm0 5.838c-3.403 0-6.162 2.759-6.162 6.162s2.759 6.163 6.162 6.163 6.162-2.759 6.162-6.163c0-3.403-2.759-6.162-6.162-6.162zm0 10.162c-2.209 0-4-1.79-4-4 0-2.209 1.791-4 4-4s4 1.791 4 4c0 2.21-1.791 4-4 4zm6.406-11.845c-.796 0-1.441.645-1.441 1.44s.645 1.44 1.441 1.44c.795 0 1.439-.645 1.439-1.44s-.644-1.44-1.439-1.44z"/>
					</svg>
				</div>
				<h3 class="font-semibold mb-2">%s</h3>
				<p class="text-sm text-gray-600 mb-3">%s</p>
				<a href="/connect/instagram?name=%s" class="inline-block bg-gradient-to-r from-purple-500 to-pink-500 text-white px-4 py-2 rounded-lg hover:from-purple-600 hover:to-pink-600 transition-colors text-sm">
					Connect Instagram
				</a>
			</div>`, provider.Name, provider.Description, provider.Name)
	}

	// Facebook providers
	for _, provider := range availableProviders["facebook"] {
		html += fmt.Sprintf(`
			<div class="border rounded-lg p-4 text-center hover:shadow-md transition-shadow">
				<div class="w-12 h-12 bg-blue-600 rounded-full mx-auto mb-3 flex items-center justify-center">
					<svg class="w-6 h-6 text-white" fill="currentColor" viewBox="0 0 24 24">
						<path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z"/>
					</svg>
				</div>
				<h3 class="font-semibold mb-2">%s</h3>
				<p class="text-sm text-gray-600 mb-3">%s</p>
				<a href="/connect/facebook?name=%s" class="inline-block bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors text-sm">
					Connect Facebook
				</a>
			</div>`, provider.Name, provider.Description, provider.Name)
	}

	html += `</div>`

	if _, err := w.Write([]byte(html)); err != nil {
		// Log error but don't fail the operation since response headers are already set
		_ = err // explicitly ignore error
	}
}

func (h *Handler) HandleDisconnect(w http.ResponseWriter, r *http.Request) {
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

	userID := h.getUserID(r)
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

func (h *Handler) getUserID(r *http.Request) string {
	return "default_user"
}
