package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tkowalski/socgo/internal/config"
	"github.com/tkowalski/socgo/internal/database"
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

func (s *Service) GetConnectURL(userID string, providerType ProviderType, providerName string) (string, error) {
	metadata, exists := SupportedProviders[providerType]
	if !exists {
		return "", fmt.Errorf("unsupported provider: %s", providerType)
	}

	// Get provider configuration
	providerConfig, err := s.config.GetProviderConfig(string(providerType), providerName)
	if err != nil {
		return "", fmt.Errorf("provider configuration not found: %s/%s", providerType, providerName)
	}

	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", providerConfig.ClientID)
	params.Add("redirect_uri", s.getRedirectURI(providerType))
	params.Add("scope", strings.Join(metadata.Scopes, " "))
	params.Add("state", fmt.Sprintf("%s:%s", userID, providerName))

	return metadata.AuthURL + "?" + params.Encode(), nil
}

func (s *Service) HandleCallback(userID string, providerType ProviderType, code string, providerName string) error {
	_, exists := SupportedProviders[providerType]
	if !exists {
		return fmt.Errorf("unsupported provider: %s", providerType)
	}

	// Get provider configuration
	providerConfig, err := s.config.GetProviderConfig(string(providerType), providerName)
	if err != nil {
		return fmt.Errorf("provider configuration not found: %s/%s", providerType, providerName)
	}

	token, err := s.exchangeCodeForToken(providerType, code, providerConfig)
	if err != nil {
		return fmt.Errorf("failed to exchange code for token: %w", err)
	}

	userInfo, err := s.getUserInfo(providerType, token.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}

	token.UserInfo = userInfo

	return s.saveProviderConfig(userID, providerType, providerName, token)
}

func (s *Service) exchangeCodeForToken(providerType ProviderType, code string, providerConfig *config.ProviderInstance) (*ProviderConfig, error) {
	metadata := SupportedProviders[providerType]

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", providerConfig.ClientID)
	data.Set("client_secret", providerConfig.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", s.getRedirectURI(providerType))

	resp, err := http.PostForm(metadata.TokenURL, data)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log error but don't fail the operation
			_ = err // explicitly ignore error
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed with status: %d", resp.StatusCode)
	}

	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)

	return &ProviderConfig{
		AccessToken:  tokenResponse.AccessToken,
		RefreshToken: tokenResponse.RefreshToken,
		TokenType:    tokenResponse.TokenType,
		ExpiresAt:    expiresAt,
		Scope:        tokenResponse.Scope,
	}, nil
}

func (s *Service) getUserInfo(providerType ProviderType, accessToken string) (*UserInfo, error) {
	metadata := SupportedProviders[providerType]

	req, err := http.NewRequest("GET", metadata.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log error but don't fail the operation
			_ = err // explicitly ignore error
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info request failed with status: %d", resp.StatusCode)
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func (s *Service) saveProviderConfig(userID string, providerType ProviderType, providerName string, config *ProviderConfig) error {
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
	result := db.Where("user_id = ? AND name = ?", userID, providerName).First(&existingProvider)

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

func (s *Service) getRedirectURI(providerType ProviderType) string {
	baseURL := s.config.Server.BaseURL
	metadata := SupportedProviders[providerType]
	return baseURL + metadata.RedirectURI
}

// GetAvailableProviders returns all available provider instances from config
func (s *Service) GetAvailableProviders() map[string][]config.ProviderInstance {
	return map[string][]config.ProviderInstance{
		"tiktok":    s.config.GetAllProviderInstances("tiktok"),
		"instagram": s.config.GetAllProviderInstances("instagram"),
		"facebook":  s.config.GetAllProviderInstances("facebook"),
	}
}
