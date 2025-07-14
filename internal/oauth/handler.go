package oauth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
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
            <div class="space-y-4">`

	if len(providers) == 0 {
		html += `<p class="text-gray-500">No providers connected yet.</p>`
	} else {
		for _, provider := range providers {
			html += fmt.Sprintf(`
                <div class="border rounded-lg p-4 flex items-center justify-between">
                    <div>
                        <h3 class="font-semibold capitalize">%s</h3>
                        <p class="text-sm text-gray-600">Connected on %s</p>
                    </div>
                    <div class="flex items-center space-x-2">
                        <span class="bg-green-100 text-green-800 px-2 py-1 rounded-full text-xs">Active</span>
                    </div>
                </div>`, provider.Name, provider.CreatedAt.Format("January 2, 2006"))
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

func (h *Handler) getUserID(r *http.Request) string {
	return "default_user"
}