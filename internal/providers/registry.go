package providers

import (
	"fmt"
	"net/http"
)

// ProviderType represents the type of social media provider
type ProviderType string

const (
	ProviderTypeTikTok    ProviderType = "tiktok"
	ProviderTypeInstagram ProviderType = "instagram"
	ProviderTypeFacebook  ProviderType = "facebook"
)

// ProviderRegistry manages provider instances
type ProviderRegistry struct {
	providers map[ProviderType]Provider
}

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[ProviderType]Provider),
	}
}

// Register registers a provider with the registry
func (r *ProviderRegistry) Register(providerType ProviderType, provider Provider) {
	r.providers[providerType] = provider
}

// Get retrieves a provider by type
func (r *ProviderRegistry) Get(providerType ProviderType) (Provider, error) {
	provider, exists := r.providers[providerType]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", providerType)
	}
	return provider, nil
}

// GetSupportedProviders returns all supported provider types
func (r *ProviderRegistry) GetSupportedProviders() []ProviderType {
	types := make([]ProviderType, 0, len(r.providers))
	for providerType := range r.providers {
		types = append(types, providerType)
	}
	return types
}

// ProviderFactory creates provider instances
type ProviderFactory struct {
	httpClient HTTPClient
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory(httpClient HTTPClient) *ProviderFactory {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return &ProviderFactory{
		httpClient: httpClient,
	}
}

// CreateProvider creates a provider instance for the given type and config
func (f *ProviderFactory) CreateProvider(providerType ProviderType, config *ProviderConfig) (Provider, error) {
	switch providerType {
	case ProviderTypeTikTok:
		return NewTikTokProvider(config, f.httpClient), nil
	case ProviderTypeInstagram:
		return NewInstagramProvider(config, f.httpClient), nil
	case ProviderTypeFacebook:
		return NewFacebookProvider(config, f.httpClient), nil
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
}

// NewTikTokProvider creates a new TikTok provider instance
func NewTikTokProvider(config *ProviderConfig, httpClient HTTPClient) Provider {
	return &TikTokProvider{
		config:     config,
		httpClient: httpClient,
	}
}

// NewInstagramProvider creates a new Instagram provider instance
func NewInstagramProvider(config *ProviderConfig, httpClient HTTPClient) Provider {
	return &InstagramProvider{
		config:     config,
		httpClient: httpClient,
	}
}

// NewFacebookProvider creates a new Facebook provider instance
func NewFacebookProvider(config *ProviderConfig, httpClient HTTPClient) Provider {
	return &FacebookProvider{
		config:     config,
		httpClient: httpClient,
	}
}

