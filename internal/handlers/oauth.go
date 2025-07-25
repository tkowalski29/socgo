package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	oauth_data "github.com/tkowalski/socgo/internal/data/oauth"
	"github.com/tkowalski/socgo/internal/handlers/internal"
	"github.com/tkowalski/socgo/internal/service/oauth"
)

type OAuthHandler struct {
	oauthService *oauth.Service
}

func NewOAuthHandler(oauthService *oauth.Service) *OAuthHandler {
	return &OAuthHandler{
		oauthService: oauthService,
	}
}

func (h *OAuthHandler) HandleConnect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider"]
	providerName := r.URL.Query().Get("name")

	if providerName == "" {
		// Redirect with error message
		errorMsg := url.QueryEscape("Provider name is required")
		http.Redirect(w, r, "/?flash="+errorMsg+"&flash_type=error", http.StatusTemporaryRedirect)
		return
	}

	providerType := oauth_data.ProviderType(strings.ToLower(provider))

	if _, exists := oauth_data.SupportedProviders[providerType]; !exists {
		// Redirect with error message
		errorMsg := url.QueryEscape(fmt.Sprintf("Unsupported provider: %s", provider))
		http.Redirect(w, r, "/?flash="+errorMsg+"&flash_type=error", http.StatusTemporaryRedirect)
		return
	}

	userID := internal.GetUserID(r)
	if userID == "" {
		// Redirect with error message
		errorMsg := url.QueryEscape("User not authenticated")
		http.Redirect(w, r, "/?flash="+errorMsg+"&flash_type=error", http.StatusTemporaryRedirect)
		return
	}

	connectURL, err := h.oauthService.GetConnectURL(userID, providerType, providerName)
	if err != nil {
		// Redirect with error message
		errorMsg := url.QueryEscape(fmt.Sprintf("Failed to generate connect URL: %v", err))
		http.Redirect(w, r, "/?flash="+errorMsg+"&flash_type=error", http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, connectURL, http.StatusTemporaryRedirect)
}

func (h *OAuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider"]

	providerType := oauth_data.ProviderType(strings.ToLower(provider))

	if _, exists := oauth_data.SupportedProviders[providerType]; !exists {
		// Redirect with error message
		errorMsg := url.QueryEscape(fmt.Sprintf("Unsupported provider: %s", provider))
		http.Redirect(w, r, "/?flash="+errorMsg+"&flash_type=error", http.StatusTemporaryRedirect)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		// Redirect with error message
		errorMsg := url.QueryEscape("Authorization code not provided")
		http.Redirect(w, r, "/?flash="+errorMsg+"&flash_type=error", http.StatusTemporaryRedirect)
		return
	}

	state := r.URL.Query().Get("state")
	if state == "" {
		// Redirect with error message
		errorMsg := url.QueryEscape("State parameter missing")
		http.Redirect(w, r, "/?flash="+errorMsg+"&flash_type=error", http.StatusTemporaryRedirect)
		return
	}

	// Parse state to get userID and providerName
	stateParts := strings.Split(state, ":")
	if len(stateParts) != 2 {
		// Redirect with error message
		errorMsg := url.QueryEscape("Invalid state parameter")
		http.Redirect(w, r, "/?flash="+errorMsg+"&flash_type=error", http.StatusTemporaryRedirect)
		return
	}

	userID := stateParts[0]
	providerName := stateParts[1]

	err := h.oauthService.HandleCallback(userID, providerType, code, providerName)
	if err != nil {
		// Redirect with error message
		errorMsg := url.QueryEscape(fmt.Sprintf("Failed to connect provider: %v", err))
		http.Redirect(w, r, "/?flash="+errorMsg+"&flash_type=error", http.StatusTemporaryRedirect)
		return
	}

	// Redirect with success message
	successMsg := url.QueryEscape(fmt.Sprintf("Successfully connected to %s (%s)", providerName, providerType))
	http.Redirect(w, r, "/?flash="+successMsg+"&flash_type=success", http.StatusTemporaryRedirect)
}
