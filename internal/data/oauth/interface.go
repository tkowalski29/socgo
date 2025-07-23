package oauth

import (
	"github.com/tkowalski/socgo/internal/data/config"
)

// ProviderAuth defines the interface for provider-specific OAuth authentication
type ProviderAuth interface {
	// GetConnectURL generates the OAuth connect URL for the provider
	GetConnectURL(userID string, providerName string, providerConfig *config.ProviderInstance) (string, error)

	// ExchangeCodeForToken exchanges authorization code for access token
	ExchangeCodeForToken(code string, providerConfig *config.ProviderInstance) (*ProviderConfig, error)

	// GetUserInfo retrieves user information using the access token
	GetUserInfo(accessToken string) (*UserInfo, error)

	// GetAvailableAccounts retrieves list of available accounts (for providers that support multiple accounts)
	GetAvailableAccounts(accessToken string) ([]AccountInfo, error)

	// GetProviderType returns the provider type
	GetProviderType() ProviderType

	// SaveAllAccounts saves all available accounts as separate providers
	SaveAllAccounts(userID string, providerName string, token *ProviderConfig, saveFunc func(userID string, providerType ProviderType, providerName string, config *ProviderConfig) error) error
}
