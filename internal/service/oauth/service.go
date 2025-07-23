package oauth

import (
	"encoding/json"
	"fmt"

	"github.com/tkowalski/socgo/internal/data/config"
	"github.com/tkowalski/socgo/internal/data/oauth"
	"github.com/tkowalski/socgo/internal/database"
	"github.com/tkowalski/socgo/internal/social/facebook"
	"github.com/tkowalski/socgo/internal/social/instagram"
	"github.com/tkowalski/socgo/internal/social/tiktok"
)

type Service struct {
	dbManager *database.Manager
	config    *config.Config
}

func NewService(dbManager *database.Manager, cfg *config.Config) *Service {
	return &Service{
		dbManager: dbManager,
		config:    cfg,
	}
}

// createProviderAuth creates a provider auth instance for the given provider type
func (s *Service) createProviderAuth(providerType oauth.ProviderType) (oauth.ProviderAuth, error) {
	switch providerType {
	case oauth.ProviderTypeInstagram:
		return instagram.NewOAuth(s.config), nil
	case oauth.ProviderTypeFacebook:
		return facebook.NewOAuth(s.config), nil
	case oauth.ProviderTypeTikTok:
		return tiktok.NewOAuth(s.config), nil
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
}

func (s *Service) GetConnectURL(userID string, providerType oauth.ProviderType, providerName string) (string, error) {
	auth, err := s.createProviderAuth(providerType)
	if err != nil {
		return "", err
	}

	providerConfig, err := s.config.GetProviderConfig(string(providerType), providerName)
	if err != nil {
		return "", fmt.Errorf("provider configuration not found: %s/%s", providerType, providerName)
	}

	return auth.GetConnectURL(userID, providerName, providerConfig)
}

func (s *Service) HandleCallback(userID string, providerType oauth.ProviderType, code string, providerName string) error {
	auth, err := s.createProviderAuth(providerType)
	if err != nil {
		return err
	}

	providerConfig, err := s.config.GetProviderConfig(string(providerType), providerName)
	if err != nil {
		return fmt.Errorf("provider configuration not found: %s/%s", providerType, providerName)
	}

	token, err := auth.ExchangeCodeForToken(code, providerConfig)
	if err != nil {
		return fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Delegate to provider-specific implementation
	return auth.SaveAllAccounts(userID, providerName, token, s.saveProviderConfig)
}

func (s *Service) saveProviderConfig(userID string, providerType oauth.ProviderType, providerName string, config *oauth.ProviderConfig) error {
	db, err := s.dbManager.GetDB(userID)
	if err != nil {
		return err
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return err
	}

	provider := &database.Provider{
		Name:     providerName,
		Type:     string(providerType),
		Config:   string(configJSON),
		UserID:   userID,
		IsActive: true,
	}

	var existingProvider database.Provider
	result := db.Where("user_id = ? AND type = ? AND name = ?", userID, string(providerType), providerName).First(&existingProvider)

	if result.Error == nil {
		provider.ID = existingProvider.ID
		return db.Save(provider).Error
	}

	return db.Create(provider).Error
}

func (s *Service) GetProviders(userID string) ([]database.Provider, error) {
	db, err := s.dbManager.GetDB(userID)
	if err != nil {
		return nil, err
	}

	var providers []database.Provider
	err = db.Where("user_id = ? AND is_active = ?", userID, true).Find(&providers).Error
	return providers, err
}

func (s *Service) DisconnectProvider(userID string, providerID uint) error {
	db, err := s.dbManager.GetDB(userID)
	if err != nil {
		return err
	}

	var provider database.Provider
	if err := db.First(&provider, providerID).Error; err != nil {
		return fmt.Errorf("provider not found: %w", err)
	}

	if provider.UserID != userID {
		return fmt.Errorf("provider does not belong to user")
	}

	// Set provider as inactive instead of deleting
	provider.IsActive = false
	return db.Save(&provider).Error
}

// GetAvailableProviders returns all available provider instances from config
func (s *Service) GetAvailableProviders() map[string][]config.ProviderInstance {
	return map[string][]config.ProviderInstance{
		"tiktok":    s.config.GetAllProviderInstances("tiktok"),
		"instagram": s.config.GetAllProviderInstances("instagram"),
		"facebook":  s.config.GetAllProviderInstances("facebook"),
	}
}
