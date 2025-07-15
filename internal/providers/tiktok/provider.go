package tiktok

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/tkowalski/socgo/internal/providers"
)

// TikTokProvider implements the Provider interface for TikTok
type TikTokProvider struct {
	config     *providers.ProviderConfig
	httpClient providers.HTTPClient
}

// NewTikTokProvider creates a new TikTok provider
func NewTikTokProvider(config *providers.ProviderConfig, httpClient providers.HTTPClient) *TikTokProvider {
	return &TikTokProvider{
		config:     config,
		httpClient: httpClient,
	}
}

// Publish publishes content to TikTok
func (p *TikTokProvider) Publish(ctx context.Context, content string) (postID string, err error) {
	// TikTok API endpoint for publishing (mock implementation)
	url := "https://open-api.tiktok.com/share/video/upload/"

	// Prepare request payload
	payload := map[string]interface{}{
		"text":         content,
		"access_token": p.config.AccessToken,
		"timestamp":    time.Now().Unix(),
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
		Data struct {
			ShareID string `json:"share_id"`
		} `json:"data"`
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error.Code != "" {
		return "", fmt.Errorf("TikTok API error: %s - %s", response.Error.Code, response.Error.Message)
	}

	// Return fake postID for mock implementation
	if response.Data.ShareID != "" {
		return response.Data.ShareID, nil
	}

	// Generate fake postID for testing
	return fmt.Sprintf("tiktok_%d", time.Now().UnixNano()), nil
}

// GetStatus retrieves the status of a published post
func (p *TikTokProvider) GetStatus(ctx context.Context, postID string) (status string, err error) {
	// TikTok API endpoint for getting post status (mock implementation)
	url := fmt.Sprintf("https://open-api.tiktok.com/video/query/?video_id=%s&access_token=%s", postID, p.config.AccessToken)

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
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error.Code != "" {
		return "", fmt.Errorf("TikTok API error: %s - %s", response.Error.Code, response.Error.Message)
	}

	// Return status or default to published for mock
	if response.Data.Status != "" {
		return response.Data.Status, nil
	}

	return string(providers.PostStatusPublished), nil
}

// RefreshToken refreshes the access token
func (p *TikTokProvider) RefreshToken(ctx context.Context) error {
	if p.config.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	// TikTok API endpoint for token refresh (mock implementation)
	url := "https://open-api.tiktok.com/oauth/refresh_token/"

	payload := map[string]interface{}{
		"client_key":    "your_client_key",
		"client_secret": "your_client_secret",
		"refresh_token": p.config.RefreshToken,
		"grant_type":    "refresh_token",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
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
		Data struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			ExpiresIn    int64  `json:"expires_in"`
		} `json:"data"`
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error.Code != "" {
		return fmt.Errorf("TikTok API error: %s - %s", response.Error.Code, response.Error.Message)
	}

	// Update config with new tokens
	if response.Data.AccessToken != "" {
		p.config.AccessToken = response.Data.AccessToken
		p.config.RefreshToken = response.Data.RefreshToken
		p.config.ExpiresAt = time.Now().Add(time.Duration(response.Data.ExpiresIn) * time.Second).Unix()
	}

	return nil
}
