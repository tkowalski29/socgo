package oauth

import (
	"github.com/tkowalski/socgo/internal/data/config"
)

// BaseProviderAuth provides common functionality for all providers
type BaseProviderAuth struct {
	config *config.Config
}

// NewBaseProviderAuth creates a new base provider auth
func NewBaseProviderAuth(cfg *config.Config) *BaseProviderAuth {
	return &BaseProviderAuth{
		config: cfg,
	}
}

// GetRedirectURI returns the redirect URI for the provider
func (b *BaseProviderAuth) GetRedirectURI(providerType ProviderType) string {
	baseURL := b.config.Server.BaseURL
	metadata := SupportedProviders[providerType]
	return baseURL + metadata.RedirectURI
}
