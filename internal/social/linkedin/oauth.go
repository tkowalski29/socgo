package linkedin

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
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

func (l *OAuth) GetProviderType() oauth.ProviderType {
	return oauth.ProviderTypeLinkedIn
}

func (l *OAuth) GetConnectURL(userID string, providerName string, providerConfig *config.ProviderInstance) (string, error) {
	metadata := oauth.SupportedProviders[oauth.ProviderTypeLinkedIn]

	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", providerConfig.ClientID)
	params.Add("redirect_uri", l.GetRedirectURI(oauth.ProviderTypeLinkedIn))
	params.Add("scope", strings.Join(metadata.Scopes, " "))
	params.Add("state", fmt.Sprintf("%s:%s", userID, providerName))

	connectURL := metadata.AuthURL + "?" + params.Encode()
	log.Printf("🔗 LinkedIn OAuth URL being generated:")
	log.Printf("   Client ID: %s", providerConfig.ClientID)
	log.Printf("   Redirect URI: %s", metadata.RedirectURI)
	log.Printf("   Scopes: %s", strings.Join(metadata.Scopes, " "))
	log.Printf("   Auth URL: %s", metadata.AuthURL)
	log.Printf("   Final URL: %s", connectURL)

	return connectURL, nil
}

func (l *OAuth) ExchangeCodeForToken(code string, providerConfig *config.ProviderInstance) (*oauth.ProviderConfig, error) {
	metadata := oauth.SupportedProviders[oauth.ProviderTypeLinkedIn]

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", providerConfig.ClientID)
	data.Set("client_secret", providerConfig.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", l.GetRedirectURI(oauth.ProviderTypeLinkedIn))

	log.Printf("🔄 Exchanging code for LinkedIn access token...")
	log.Printf("   Token URL: %s", metadata.TokenURL)
	log.Printf("   Code: %s...", code[:min(len(code), 10)])

	resp, err := http.Post(metadata.TokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	log.Printf("LinkedIn token exchange response status: %d", resp.StatusCode)
	log.Printf("LinkedIn token exchange response: %s", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status: %d, body: %s", resp.StatusCode, string(body))
	}

	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
		RefreshToken string `json:"refresh_token,omitempty"`
		Scope        string `json:"scope"`
		Error        string `json:"error,omitempty"`
		ErrorDesc    string `json:"error_description,omitempty"`
	}

	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	if tokenResponse.Error != "" {
		return nil, fmt.Errorf("LinkedIn OAuth error: %s - %s", tokenResponse.Error, tokenResponse.ErrorDesc)
	}

	config := &oauth.ProviderConfig{
		AccessToken:  tokenResponse.AccessToken,
		RefreshToken: tokenResponse.RefreshToken,
		TokenType:    tokenResponse.TokenType,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
		Scope:        tokenResponse.Scope,
	}

	log.Printf("✅ LinkedIn token exchange successful")
	log.Printf("   Token Type: %s", config.TokenType)
	log.Printf("   Expires At: %s", config.ExpiresAt.Format(time.RFC3339))
	log.Printf("   Scope: %s", config.Scope)

	return config, nil
}

func (l *OAuth) GetUserInfo(accessToken string) (*oauth.UserInfo, error) {
	metadata := oauth.SupportedProviders[oauth.ProviderTypeLinkedIn]

	log.Printf("🔍 Getting LinkedIn user info...")
	log.Printf("   User Info URL: %s", metadata.UserInfoURL)

	req, err := http.NewRequest("GET", metadata.UserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user info request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read user info response: %w", err)
	}

	log.Printf("LinkedIn user info response status: %d", resp.StatusCode)
	log.Printf("LinkedIn user info response: %s", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info request failed with status: %d, body: %s", resp.StatusCode, string(body))
	}

	var userInfoResponse struct {
		Sub           string `json:"sub"`
		Name          string `json:"name"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Picture       string `json:"picture"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Error         string `json:"error,omitempty"`
		ErrorDesc     string `json:"error_description,omitempty"`
	}

	if err := json.Unmarshal(body, &userInfoResponse); err != nil {
		return nil, fmt.Errorf("failed to decode user info response: %w", err)
	}

	if userInfoResponse.Error != "" {
		return nil, fmt.Errorf("LinkedIn user info error: %s - %s", userInfoResponse.Error, userInfoResponse.ErrorDesc)
	}

	userInfo := &oauth.UserInfo{
		ID:       userInfoResponse.Sub,
		Name:     userInfoResponse.Name,
		Username: "", // LinkedIn doesn't provide username in userinfo endpoint
		Email:    userInfoResponse.Email,
		Avatar:   userInfoResponse.Picture,
	}

	log.Printf("✅ LinkedIn user info retrieved successfully")
	log.Printf("   User ID: %s", userInfo.ID)
	log.Printf("   Name: %s", userInfo.Name)
	log.Printf("   Email: %s", userInfo.Email)

	return userInfo, nil
}

func (l *OAuth) GetAvailableAccounts(accessToken string) ([]oauth.AccountInfo, error) {
	// For LinkedIn, the user account is the main account for posting
	// We'll get user info and return it as the account
	userInfo, err := l.GetUserInfo(accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info for account: %w", err)
	}

	accounts := []oauth.AccountInfo{
		{
			ID:       userInfo.ID,
			Name:     userInfo.Name,
			Username: userInfo.Username,
			Type:     "personal",
			Category: "LinkedIn Profile",
		},
	}

	log.Printf("✅ LinkedIn available accounts:")
	for _, account := range accounts {
		log.Printf("   - ID: %s, Name: %s, Type: %s", account.ID, account.Name, account.Type)
	}

	return accounts, nil
}

// SaveAllAccounts saves all available LinkedIn accounts as separate providers
func (l *OAuth) SaveAllAccounts(userID string, providerName string, token *oauth.ProviderConfig, saveFunc func(userID string, providerType oauth.ProviderType, providerName string, config *oauth.ProviderConfig) error) error {
	accounts, err := l.GetAvailableAccounts(token.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to get available LinkedIn accounts: %w", err)
	}

	if len(accounts) == 0 {
		return fmt.Errorf("no LinkedIn accounts available")
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
				ID:       account.ID,
				Name:     account.Name,
				Username: account.Username,
			},
		}

		accountProviderName := fmt.Sprintf("%s_%s", providerName, account.Name)
		log.Printf("💾 Saving LinkedIn account: %s (ID: %s)", accountProviderName, account.ID)

		if err := saveFunc(userID, oauth.ProviderTypeLinkedIn, accountProviderName, accountToken); err != nil {
			return fmt.Errorf("failed to save LinkedIn account %s: %w", accountProviderName, err)
		}
	}

	log.Printf("✅ Successfully saved %d LinkedIn accounts", len(accounts))
	return nil
}

// Helper function for min operation
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}