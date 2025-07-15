package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/tkowalski/socgo/internal/database"
	"github.com/tkowalski/socgo/internal/oauth"
)

// ProviderService manages social media providers with Strategy pattern
type ProviderService struct {
	registry     *ProviderRegistry
	factory      *ProviderFactory
	dbManager    *database.Manager
	oauthService *oauth.Service
}

// NewProviderService creates a new provider service
func NewProviderService(dbManager *database.Manager, oauthService *oauth.Service) *ProviderService {
	registry := NewProviderRegistry()
	factory := NewProviderFactory(&http.Client{})

	return &ProviderService{
		registry:     registry,
		factory:      factory,
		dbManager:    dbManager,
		oauthService: oauthService,
	}
}

// PublishContent publishes content to a specific provider
func (s *ProviderService) PublishContent(ctx context.Context, userID string, providerName string, content string) (postID string, err error) {
	// Get provider configuration from database
	config, err := s.getProviderConfig(ctx, userID, providerName)
	if err != nil {
		return "", fmt.Errorf("failed to get provider config: %w", err)
	}

	// Convert provider name to type
	providerType := ProviderType(providerName)

	// Create provider instance using factory
	provider, err := s.factory.CreateProvider(providerType, config)
	if err != nil {
		return "", fmt.Errorf("failed to create provider: %w", err)
	}

	// Publish content using provider
	postID, err = provider.Publish(ctx, content)
	if err != nil {
		return "", fmt.Errorf("failed to publish content: %w", err)
	}

	return postID, nil
}

// GetPostStatus retrieves the status of a published post
func (s *ProviderService) GetPostStatus(ctx context.Context, userID string, providerName string, postID string) (status string, err error) {
	// Get provider configuration from database
	config, err := s.getProviderConfig(ctx, userID, providerName)
	if err != nil {
		return "", fmt.Errorf("failed to get provider config: %w", err)
	}

	// Convert provider name to type
	providerType := ProviderType(providerName)

	// Create provider instance using factory
	provider, err := s.factory.CreateProvider(providerType, config)
	if err != nil {
		return "", fmt.Errorf("failed to create provider: %w", err)
	}

	// Get post status using provider
	status, err = provider.GetStatus(ctx, postID)
	if err != nil {
		return "", fmt.Errorf("failed to get post status: %w", err)
	}

	return status, nil
}

// RefreshProviderToken refreshes the access token for a provider
func (s *ProviderService) RefreshProviderToken(ctx context.Context, userID string, providerName string) error {
	// Get provider configuration from database
	config, err := s.getProviderConfig(ctx, userID, providerName)
	if err != nil {
		return fmt.Errorf("failed to get provider config: %w", err)
	}

	// Convert provider name to type
	providerType := ProviderType(providerName)

	// Create provider instance using factory
	provider, err := s.factory.CreateProvider(providerType, config)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Refresh token using provider
	err = provider.RefreshToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update provider configuration in database
	err = s.updateProviderConfig(ctx, userID, providerName, config)
	if err != nil {
		return fmt.Errorf("failed to update provider config: %w", err)
	}

	return nil
}

// GetSupportedProviders returns all supported provider types
func (s *ProviderService) GetSupportedProviders() []string {
	types := s.registry.GetSupportedProviders()
	providers := make([]string, len(types))
	for i, providerType := range types {
		providers[i] = string(providerType)
	}
	return providers
}

// IsProviderConfigured checks if a provider is configured for a user
func (s *ProviderService) IsProviderConfigured(userID string, providerName string) (bool, error) {
	db, err := s.dbManager.GetDB(userID)
	if err != nil {
		return false, fmt.Errorf("failed to get database: %w", err)
	}

	var provider database.Provider
	result := db.Where("user_id = ? AND name = ? AND is_active = ?", userID, providerName, true).First(&provider)
	if result.Error != nil {
		return false, nil // Provider not found or not active
	}

	return true, nil
}

// getProviderConfig retrieves provider configuration from database
func (s *ProviderService) getProviderConfig(ctx context.Context, userID string, providerName string) (*ProviderConfig, error) {
	db, err := s.dbManager.GetDB(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	var dbProvider database.Provider
	result := db.Where("user_id = ? AND name = ? AND is_active = ?", userID, providerName, true).First(&dbProvider)
	if result.Error != nil {
		return nil, fmt.Errorf("provider not found: %w", result.Error)
	}

	// Parse OAuth configuration
	var oauthConfig oauth.ProviderConfig
	if err := json.Unmarshal([]byte(dbProvider.Config), &oauthConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OAuth config: %w", err)
	}

	// Convert to provider config
	config := &ProviderConfig{
		AccessToken:  oauthConfig.AccessToken,
		RefreshToken: oauthConfig.RefreshToken,
		TokenType:    oauthConfig.TokenType,
		ExpiresAt:    oauthConfig.ExpiresAt.Unix(),
		Scope:        oauthConfig.Scope,
		UserID:       userID,
	}

	// Set user info if available
	if oauthConfig.UserInfo != nil {
		config.UserID = oauthConfig.UserInfo.ID
	}

	return config, nil
}

// updateProviderConfig updates provider configuration in database
func (s *ProviderService) updateProviderConfig(ctx context.Context, userID string, providerName string, config *ProviderConfig) error {
	db, err := s.dbManager.GetDB(userID)
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	var dbProvider database.Provider
	result := db.Where("user_id = ? AND name = ? AND is_active = ?", userID, providerName, true).First(&dbProvider)
	if result.Error != nil {
		return fmt.Errorf("provider not found: %w", result.Error)
	}

	// Parse existing OAuth configuration
	var oauthConfig oauth.ProviderConfig
	if err := json.Unmarshal([]byte(dbProvider.Config), &oauthConfig); err != nil {
		return fmt.Errorf("failed to unmarshal OAuth config: %w", err)
	}

	// Update OAuth configuration with new token information
	oauthConfig.AccessToken = config.AccessToken
	oauthConfig.RefreshToken = config.RefreshToken
	oauthConfig.TokenType = config.TokenType
	oauthConfig.ExpiresAt = time.Unix(config.ExpiresAt, 0)
	oauthConfig.Scope = config.Scope

	// Marshal updated configuration
	updatedConfig, err := json.Marshal(oauthConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal updated OAuth config: %w", err)
	}

	// Update database record
	dbProvider.Config = string(updatedConfig)
	dbProvider.UpdatedAt = time.Now()

	if err := db.Save(&dbProvider).Error; err != nil {
		return fmt.Errorf("failed to save provider config: %w", err)
	}

	return nil
}
