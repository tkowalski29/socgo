package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
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

	providerType := ProviderType(strings.ToLower(provider))

	if _, exists := SupportedProviders[providerType]; !exists {
		http.Error(w, fmt.Sprintf("Unsupported provider: %s", provider), http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)
	if userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	connectURL, err := h.oauthService.GetConnectURL(userID, providerType)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate connect URL: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, connectURL, http.StatusTemporaryRedirect)
}

func (h *Handler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider"]

	providerType := ProviderType(strings.ToLower(provider))

	if _, exists := SupportedProviders[providerType]; !exists {
		http.Error(w, fmt.Sprintf("Unsupported provider: %s", provider), http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Authorization code not provided", http.StatusBadRequest)
		return
	}

	state := r.URL.Query().Get("state")
	if state == "" {
		http.Error(w, "State parameter missing", http.StatusBadRequest)
		return
	}

	userID := state

	err := h.oauthService.HandleCallback(userID, providerType, code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to handle callback: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/providers", http.StatusTemporaryRedirect)
}

func (h *Handler) HandleProviders(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	if userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
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

func (h *Handler) handleProvidersJSON(w http.ResponseWriter, providers []database.Provider) {
	type ProviderResponse struct {
		DisplayName string `json:"display_name"`
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

	html := `<!DOCTYPE html>
<html>
<head>
    <title>Connected Providers</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100 min-h-screen py-8">
    <div class="max-w-4xl mx-auto px-4">
        <h1 class="text-3xl font-bold text-gray-900 mb-8">Connected Providers</h1>
        
        <div class="bg-white rounded-lg shadow p-6 mb-8">
            <h2 class="text-xl font-semibold mb-4">Connect New Provider</h2>
            <div class="space-y-4">
                <a href="/connect/tiktok" class="inline-block bg-black text-white px-4 py-2 rounded hover:bg-gray-800 transition">
                    Connect TikTok
                </a>
                <a href="/connect/instagram" class="inline-block bg-pink-600 text-white px-4 py-2 rounded hover:bg-pink-700 transition ml-4">
                    Connect Instagram
                </a>
                <a href="/connect/facebook" class="inline-block bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 transition ml-4">
                    Connect Facebook
                </a>
            </div>
        </div>
        
        <div class="bg-white rounded-lg shadow p-6">
            <h2 class="text-xl font-semibold mb-4">Connected Accounts</h2>
            <div class="space-y-4" id="providers-list">`

	if len(providers) == 0 {
		html += `<p class="text-gray-500">No providers connected yet.</p>`
	} else {
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

			html += fmt.Sprintf(`
                <div class="border rounded-lg p-4 flex items-center justify-between">
                    <div>
                        <h3 class="font-semibold capitalize">%s</h3>
                        <p class="text-sm text-gray-600">Connected on %s</p>
                    </div>
                    <div class="flex items-center space-x-2">
                        <span class="%s px-2 py-1 rounded-full text-xs">%s</span>
                        <button hx-delete="/providers/%d" hx-target="closest div" hx-swap="outerHTML" hx-confirm="Are you sure you want to disconnect this provider?" class="bg-red-500 hover:bg-red-700 text-white font-bold py-1 px-3 rounded text-sm">
                            Disconnect
                        </button>
                    </div>
                </div>`, displayName, provider.CreatedAt.Format("January 2, 2006"), statusClass, status, provider.ID)
		}
	}

	html += `
            </div>
        </div>
    </div>
</body>
</html>`

	_, _ = w.Write([]byte(html))
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

	// Return empty response for HTMX to remove the element
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) getUserID(r *http.Request) string {
	return "default_user"
}
