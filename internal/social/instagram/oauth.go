package instagram

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

func (i *OAuth) GetProviderType() oauth.ProviderType {
	return oauth.ProviderTypeInstagram
}

func (i *OAuth) GetConnectURL(userID string, providerName string, providerConfig *config.ProviderInstance) (string, error) {
	metadata := oauth.SupportedProviders[oauth.ProviderTypeInstagram]

	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", providerConfig.ClientID)
	params.Add("redirect_uri", i.GetRedirectURI(oauth.ProviderTypeInstagram))
	params.Add("scope", strings.Join(metadata.Scopes, " "))
	params.Add("state", fmt.Sprintf("%s:%s", userID, providerName))

	return metadata.AuthURL + "?" + params.Encode(), nil
}

func (i *OAuth) ExchangeCodeForToken(code string, providerConfig *config.ProviderInstance) (*oauth.ProviderConfig, error) {
	metadata := oauth.SupportedProviders[oauth.ProviderTypeInstagram]

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", providerConfig.ClientID)
	data.Set("client_secret", providerConfig.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", i.GetRedirectURI(oauth.ProviderTypeInstagram))

	resp, err := http.PostForm(metadata.TokenURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token request failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)

	return &oauth.ProviderConfig{
		AccessToken: tokenResponse.AccessToken,
		TokenType:   tokenResponse.TokenType,
		ExpiresAt:   expiresAt,
	}, nil
}

func (i *OAuth) GetUserInfo(accessToken string) (*oauth.UserInfo, error) {
	req, err := http.NewRequest("GET", "https://graph.facebook.com/v18.0/me", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("access_token", accessToken)
	q.Add("fields", "id,name")
	req.URL.RawQuery = q.Encode()

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

	var userInfo oauth.UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info response: %w", err)
	}

	return &userInfo, nil
}

func (i *OAuth) GetAvailableAccounts(accessToken string) ([]oauth.AccountInfo, error) {
	req, err := http.NewRequest("GET", "https://graph.facebook.com/v18.0/me/accounts", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("access_token", accessToken)
	q.Add("fields", "id,name,category,instagram_business_account")
	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Instagram accounts request failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		Data []struct {
			ID                       string `json:"id"`
			Name                     string `json:"name"`
			Category                 string `json:"category"`
			InstagramBusinessAccount struct {
				ID string `json:"id"`
			} `json:"instagram_business_account"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode Instagram accounts response: %w", err)
	}

	var accounts []oauth.AccountInfo
	for _, page := range response.Data {
		if page.InstagramBusinessAccount.ID != "" {
			accounts = append(accounts, oauth.AccountInfo{
				ID:       page.InstagramBusinessAccount.ID,
				Name:     page.Name,
				Type:     "instagram",
				Category: page.Category,
			})
		}
	}

	// If no Instagram Business Accounts found, try to get the user's basic Instagram account
	if len(accounts) == 0 {
		userInfo, err := i.GetUserInfo(accessToken)
		if err != nil {
			return nil, fmt.Errorf("failed to get user info for basic Instagram account: %w", err)
		}

		// For basic Instagram accounts, we use the user's ID and name
		accounts = append(accounts, oauth.AccountInfo{
			ID:       userInfo.ID,
			Name:     userInfo.Name,
			Type:     "instagram_basic",
			Category: "Personal Account",
		})
	}

	return accounts, nil
}

// SaveAllAccounts saves all available Instagram accounts as separate providers
func (i *OAuth) SaveAllAccounts(userID string, providerName string, token *oauth.ProviderConfig, saveFunc func(userID string, providerType oauth.ProviderType, providerName string, config *oauth.ProviderConfig) error) error {
	accounts, err := i.GetAvailableAccounts(token.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to get available accounts: %w", err)
	}

	if len(accounts) == 0 {
		return fmt.Errorf("no Instagram accounts available")
	}

	// Save each account as a separate provider
	for _, account := range accounts {
		accountToken := &oauth.ProviderConfig{
			AccessToken: token.AccessToken,
			TokenType:   token.TokenType,
			ExpiresAt:   token.ExpiresAt,
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

		if err := saveFunc(userID, i.GetProviderType(), accountProviderName, accountToken); err != nil {
			return fmt.Errorf("failed to save Instagram account %s: %w", account.ID, err)
		}
	}

	return nil
}
