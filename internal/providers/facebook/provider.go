package facebook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/tkowalski/socgo/internal/providers"
)

// FacebookProvider implements the Provider interface for Facebook
type FacebookProvider struct {
	config     *providers.ProviderConfig
	httpClient providers.HTTPClient
}

// NewFacebookProvider creates a new Facebook provider
func NewFacebookProvider(config *providers.ProviderConfig, httpClient providers.HTTPClient) *FacebookProvider {
	return &FacebookProvider{
		config:     config,
		httpClient: httpClient,
	}
}

// Publish publishes content to Facebook
func (p *FacebookProvider) Publish(ctx context.Context, content string) (postID string, err error) {
	// Facebook Graph API endpoint for publishing (mock implementation)
	url := fmt.Sprintf("https://graph.facebook.com/%s/feed", p.config.UserID)

	// Prepare request payload
	payload := map[string]interface{}{
		"message":      content,
		"access_token": p.config.AccessToken,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.AccessToken)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log error but don't fail the operation
				_ = err // explicitly ignore error
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var response struct {
		ID    string `json:"id"`
		Error struct {
			Message   string `json:"message"`
			Type      string `json:"type"`
			Code      int    `json:"code"`
			ErrorCode int    `json:"error_code"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error.Message != "" {
		return "", fmt.Errorf("Facebook API error: %s (type: %s, code: %d)",
			response.Error.Message, response.Error.Type, response.Error.Code)
	}

	// Return fake postID for mock implementation
	if response.ID != "" {
		return response.ID, nil
	}

	// Generate fake postID for testing
	return fmt.Sprintf("facebook_%d", time.Now().UnixNano()), nil
}

// GetStatus retrieves the status of a published post
func (p *FacebookProvider) GetStatus(ctx context.Context, postID string) (status string, err error) {
	// Facebook Graph API endpoint for getting post status (mock implementation)
	url := fmt.Sprintf("https://graph.facebook.com/%s?fields=id,message,created_time,status_type&access_token=%s",
		postID, p.config.AccessToken)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.config.AccessToken)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log error but don't fail the operation
				_ = err // explicitly ignore error
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var response struct {
		ID          string `json:"id"`
		Message     string `json:"message"`
		CreatedTime string `json:"created_time"`
		StatusType  string `json:"status_type"`
		Error       struct {
			Message   string `json:"message"`
			Type      string `json:"type"`
			Code      int    `json:"code"`
			ErrorCode int    `json:"error_code"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error.Message != "" {
		return "", fmt.Errorf("Facebook API error: %s (type: %s, code: %d)",
			response.Error.Message, response.Error.Type, response.Error.Code)
	}

	// Return status based on response or default to published for mock
	if response.ID != "" && response.CreatedTime != "" {
		return string(providers.PostStatusPublished), nil
	}

	return string(providers.PostStatusPending), nil
}

// RefreshToken refreshes the access token
func (p *FacebookProvider) RefreshToken(ctx context.Context) error {
	if p.config.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	// Facebook Graph API endpoint for token refresh (mock implementation)
	url := "https://graph.facebook.com/oauth/access_token"

	payload := map[string]interface{}{
		"grant_type":        "fb_exchange_token",
		"client_id":         "your_app_id",
		"client_secret":     "your_app_secret",
		"fb_exchange_token": p.config.AccessToken,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log error but don't fail the operation
				_ = err // explicitly ignore error
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token refresh failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var response struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int64  `json:"expires_in"`
		Error       struct {
			Message   string `json:"message"`
			Type      string `json:"type"`
			Code      int    `json:"code"`
			ErrorCode int    `json:"error_code"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error.Message != "" {
		return fmt.Errorf("Facebook API error: %s (type: %s, code: %d)",
			response.Error.Message, response.Error.Type, response.Error.Code)
	}

	// Update config with new tokens
	if response.AccessToken != "" {
		p.config.AccessToken = response.AccessToken
		p.config.TokenType = response.TokenType
		p.config.ExpiresAt = time.Now().Add(time.Duration(response.ExpiresIn) * time.Second).Unix()
	}

	return nil
}
