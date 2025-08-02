package post

import (
	"fmt"
	"net/http"

	"github.com/tkowalski/socgo/internal/data/provider"
	"github.com/tkowalski/socgo/internal/social/facebook"
	"github.com/tkowalski/socgo/internal/social/instagram"
	"github.com/tkowalski/socgo/internal/social/linkedin"
	"github.com/tkowalski/socgo/internal/social/tiktok"
)

// ProviderType represents the type of social media provider
type ProviderType string

const (
	ProviderTypeTikTok    ProviderType = "tiktok"
	ProviderTypeInstagram ProviderType = "instagram"
	ProviderTypeFacebook  ProviderType = "facebook"
	ProviderTypeLinkedIn  ProviderType = "linkedin"
)

// ProviderRegistry manages provider instances
type ProviderRegistry struct {
	providers map[ProviderType]provider.Provider
}

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[ProviderType]provider.Provider),
	}
}

// Register registers a provider with the registry
func (r *ProviderRegistry) Register(providerType ProviderType, provider provider.Provider) {
	r.providers[providerType] = provider
}

// Get retrieves a provider by type
func (r *ProviderRegistry) Get(providerType ProviderType) (provider.Provider, error) {
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
	httpClient provider.HTTPClient
	baseURL    string
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory(httpClient provider.HTTPClient, baseURL string) *ProviderFactory {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return &ProviderFactory{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// CreateProvider creates a provider instance for the given type and config
func (f *ProviderFactory) CreateProvider(providerType ProviderType, config *provider.ProviderConfig) (provider.Provider, error) {
	switch providerType {
	case ProviderTypeTikTok:
		return NewTikTokProvider(config, f.httpClient), nil
	case ProviderTypeInstagram:
		return NewInstagramProvider(config, f.httpClient, f.baseURL), nil
	case ProviderTypeFacebook:
		return NewFacebookProvider(config, f.httpClient), nil
	case ProviderTypeLinkedIn:
		return NewLinkedInProvider(config, f.httpClient, f.baseURL), nil
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
}

// NewTikTokProvider creates a new TikTok provider instance
func NewTikTokProvider(config *provider.ProviderConfig, httpClient provider.HTTPClient) provider.Provider {
	return &tiktok.TikTokPost{
		Config:     config,
		HttpClient: httpClient,
	}
}

// NewInstagramProvider creates a new Instagram provider instance
func NewInstagramProvider(config *provider.ProviderConfig, httpClient provider.HTTPClient, baseURL string) provider.Provider {
	return instagram.NewInstagramPost(config, httpClient, baseURL)
}

// NewFacebookProvider creates a new Facebook provider instance
func NewFacebookProvider(config *provider.ProviderConfig, httpClient provider.HTTPClient) provider.Provider {
	return &facebook.FacebookPost{
		Config:     config,
		HttpClient: httpClient,
	}
}

// NewLinkedInProvider creates a new LinkedIn provider instance
func NewLinkedInProvider(config *provider.ProviderConfig, httpClient provider.HTTPClient, baseURL string) provider.Provider {
	return linkedin.NewLinkedInPost(config, httpClient, baseURL)
}
