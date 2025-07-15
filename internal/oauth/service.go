package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tkowalski/socgo/internal/database"
)

type Service struct {
	dbManager *database.Manager
}

func NewService(dbManager *database.Manager) *Service {
	return &Service{
		dbManager: dbManager,
	}
}

func (s *Service) GetConnectURL(userID string, providerType ProviderType) (string, error) {
	metadata, exists := SupportedProviders[providerType]
	if !exists {
		return "", fmt.Errorf("unsupported provider: %s", providerType)
	}

	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", s.getClientID(providerType))
	params.Add("redirect_uri", s.getRedirectURI(providerType))
	params.Add("scope", strings.Join(metadata.Scopes, " "))
	params.Add("state", userID)

	return metadata.AuthURL + "?" + params.Encode(), nil
}

func (s *Service) HandleCallback(userID string, providerType ProviderType, code string) error {
	_, exists := SupportedProviders[providerType]
	if !exists {
		return fmt.Errorf("unsupported provider: %s", providerType)
	}

	token, err := s.exchangeCodeForToken(providerType, code)
	if err != nil {
		return fmt.Errorf("failed to exchange code for token: %w", err)
	}

	userInfo, err := s.getUserInfo(providerType, token.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}

	token.UserInfo = userInfo

	return s.saveProviderConfig(userID, providerType, token)
}

func (s *Service) exchangeCodeForToken(providerType ProviderType, code string) (*ProviderConfig, error) {
	metadata := SupportedProviders[providerType]

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", s.getClientID(providerType))
	data.Set("client_secret", s.getClientSecret(providerType))
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

func (s *Service) saveProviderConfig(userID string, providerType ProviderType, config *ProviderConfig) error {
	db, err := s.dbManager.GetDB(userID)
	if err != nil {
		return err
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return err
	}

	provider := &database.Provider{
		Name:     string(providerType),
		Type:     string(providerType),
		Config:   string(configJSON),
		UserID:   userID,
		IsActive: true,
	}

	var existingProvider database.Provider
	result := db.Where("user_id = ? AND type = ?", userID, string(providerType)).First(&existingProvider)

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

func (s *Service) getClientID(providerType ProviderType) string {
	switch providerType {
	case ProviderTypeTikTok:
		return "your_tiktok_client_id"
	case ProviderTypeInstagram:
		return "your_instagram_client_id"
	case ProviderTypeFacebook:
		return "your_facebook_client_id"
	default:
		return ""
	}
}

func (s *Service) getClientSecret(providerType ProviderType) string {
	switch providerType {
	case ProviderTypeTikTok:
		return "your_tiktok_client_secret"
	case ProviderTypeInstagram:
		return "your_instagram_client_secret"
	case ProviderTypeFacebook:
		return "your_facebook_client_secret"
	default:
		return ""
	}
}

func (s *Service) getRedirectURI(providerType ProviderType) string {
	baseURL := "http://localhost:8080"
	metadata := SupportedProviders[providerType]
	return baseURL + metadata.RedirectURI
}
