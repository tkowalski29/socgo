package tiktok

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tkowalski/socgo/internal/data/config"
	"github.com/tkowalski/socgo/internal/data/oauth"
)

type OAuth struct {
	*oauth.BaseProviderAuth
}

func NewOAuth(cfg *config.Config) *OAuth {
	return &OAuth{
		BaseProviderAuth: oauth.NewBaseProviderAuth(cfg),
	}
}

func (t *OAuth) GetProviderType() oauth.ProviderType {
	return oauth.ProviderTypeTikTok
}

func (t *OAuth) GetConnectURL(userID string, providerName string, providerConfig *config.ProviderInstance) (string, error) {
	metadata := oauth.SupportedProviders[oauth.ProviderTypeTikTok]

	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_key", providerConfig.ClientID) // TikTok requires client_key
	params.Add("redirect_uri", t.GetRedirectURI(oauth.ProviderTypeTikTok))
	params.Add("scope", strings.Join(metadata.Scopes, " "))
	params.Add("state", fmt.Sprintf("%s:%s", userID, providerName))
	params.Add("auth_type", "code") // TikTok specific parameter

	return metadata.AuthURL + "?" + params.Encode(), nil
}

func (t *OAuth) ExchangeCodeForToken(code string, providerConfig *config.ProviderInstance) (*oauth.ProviderConfig, error) {
	metadata := oauth.SupportedProviders[oauth.ProviderTypeTikTok]

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_key", providerConfig.ClientID) // TikTok uses client_key instead of client_id
	data.Set("client_secret", providerConfig.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", t.GetRedirectURI(oauth.ProviderTypeTikTok))

	req, err := http.NewRequest("POST", metadata.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token request failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)

	return &oauth.ProviderConfig{
		AccessToken:  tokenResponse.AccessToken,
		RefreshToken: tokenResponse.RefreshToken,
		TokenType:    tokenResponse.TokenType,
		ExpiresAt:    expiresAt,
		Scope:        tokenResponse.Scope,
	}, nil
}

func (t *OAuth) GetUserInfo(accessToken string) (*oauth.UserInfo, error) {
	// Use original TikTok endpoint with fields parameter
	req, err := http.NewRequest("GET", "https://open.tiktokapis.com/v2/user/info/?fields=open_id,display_name,avatar_url", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("user info request failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		Data struct {
			User struct {
				OpenID      string `json:"open_id"`
				DisplayName string `json:"display_name"`
				AvatarURL   string `json:"avatar_url"`
			} `json:"user"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode user info response: %w", err)
	}

	return &oauth.UserInfo{
		ID:       response.Data.User.OpenID,
		Name:     response.Data.User.DisplayName,
		Username: response.Data.User.DisplayName, // TikTok doesn't have username field
		Avatar:   response.Data.User.AvatarURL,
	}, nil
}

func (t *OAuth) GetAvailableAccounts(accessToken string) ([]oauth.AccountInfo, error) {
	// TikTok doesn't support multiple accounts in the same way as Facebook/Instagram
	// So we just return the user info as a single account
	userInfo, err := t.GetUserInfo(accessToken)
	if err != nil {
		return nil, err
	}

	return []oauth.AccountInfo{
		{
			ID:       userInfo.ID,
			Name:     userInfo.Name,
			Username: userInfo.Username,
			Type:     "user",
		},
	}, nil
}

// SaveAllAccounts saves all available TikTok accounts as separate providers
func (t *OAuth) SaveAllAccounts(userID string, providerName string, token *oauth.ProviderConfig, saveFunc func(userID string, providerType oauth.ProviderType, providerName string, config *oauth.ProviderConfig) error) error {
	accounts, err := t.GetAvailableAccounts(token.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to get available accounts: %w", err)
	}

	if len(accounts) == 0 {
		return fmt.Errorf("no TikTok accounts available")
	}

	// Save each account as a separate provider
	for _, account := range accounts {
		accountToken := &oauth.ProviderConfig{
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			TokenType:    token.TokenType,
			ExpiresAt:    token.ExpiresAt,
			Scope:        token.Scope,
			UserInfo: &oauth.UserInfo{
				ID:   account.ID,
				Name: account.Name,
			},
		}

		// Use account name as provider name
		accountProviderName := account.Name
		if accountProviderName == "" {
			accountProviderName = fmt.Sprintf("%s_%s", providerName, account.ID)
		}

		if err := saveFunc(userID, t.GetProviderType(), accountProviderName, accountToken); err != nil {
			return fmt.Errorf("failed to save TikTok account %s: %w", account.ID, err)
		}
	}

	return nil
}
