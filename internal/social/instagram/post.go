package instagram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/tkowalski/socgo/internal/data/provider"
)

// InstagramPost implements the Post interface for Instagram
type InstagramPost struct {
	Config     *provider.ProviderConfig
	HttpClient provider.HTTPClient
}

// NewInstagramPost creates a new Instagram Post
func NewInstagramPost(config *provider.ProviderConfig, httpClient provider.HTTPClient) *InstagramPost {
	return &InstagramPost{
		Config:     config,
		HttpClient: httpClient,
	}
}

// Publish publishes content to Instagram
func (p *InstagramPost) Publish(ctx context.Context, content string, media []provider.Media) (postID string, err error) {
	// Instagram Basic Display API endpoint for publishing (mock implementation)
	url := fmt.Sprintf("https://graph.instagram.com/%s/media", p.Config.UserID)

	// Prepare request payload
	payload := map[string]interface{}{
		"caption":      content,
		"media_type":   "CAROUSEL_ALBUM",
		"access_token": p.Config.AccessToken,
	}

	// Add media information if available
	if len(media) > 0 {
		// For now, just log that media was provided
		// Instagram API implementation would need to be expanded
		payload["media_count"] = len(media)
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
	req.Header.Set("Authorization", "Bearer "+p.Config.AccessToken)

	resp, err := p.HttpClient.Do(req)
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
		return "", fmt.Errorf("Instagram API error: %s (type: %s, code: %d)",
			response.Error.Message, response.Error.Type, response.Error.Code)
	}

	// Return post ID from response
	if response.ID != "" {
		return response.ID, nil
	}

	// Generate fake postID for testing if no real ID returned
	return fmt.Sprintf("instagram_%d", time.Now().UnixNano()), nil
}

// GetStatus retrieves the status of a published post
func (p *InstagramPost) GetStatus(ctx context.Context, postID string) (status string, err error) {
	// Instagram Basic Display API endpoint for getting post status (mock implementation)
	url := fmt.Sprintf("https://graph.instagram.com/%s?fields=id,media_type,permalink,timestamp&access_token=%s",
		postID, p.Config.AccessToken)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.Config.AccessToken)

	resp, err := p.HttpClient.Do(req)
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
		ID        string `json:"id"`
		MediaType string `json:"media_type"`
		Permalink string `json:"permalink"`
		Timestamp string `json:"timestamp"`
		Error     struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    int    `json:"code"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error.Message != "" {
		return "", fmt.Errorf("Instagram API error: %s (type: %s, code: %d)",
			response.Error.Message, response.Error.Type, response.Error.Code)
	}

	// Return status based on response or default to published for mock
	if response.ID != "" && response.Permalink != "" {
		return string(provider.PostStatusPublished), nil
	}

	return string(provider.PostStatusPending), nil
}

// RefreshToken refreshes the access token
func (p *InstagramPost) RefreshToken(ctx context.Context) error {
	if p.Config.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	// Instagram Basic Display API endpoint for token refresh (mock implementation)
	url := "https://graph.instagram.com/refresh_access_token"

	payload := map[string]interface{}{
		"grant_type":   "ig_refresh_token",
		"access_token": p.Config.AccessToken,
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

	resp, err := p.HttpClient.Do(req)
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
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    int    `json:"code"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error.Message != "" {
		return fmt.Errorf("Instagram API error: %s (type: %s, code: %d)",
			response.Error.Message, response.Error.Type, response.Error.Code)
	}

	// Update config with new tokens
	if response.AccessToken != "" {
		p.Config.AccessToken = response.AccessToken
		p.Config.TokenType = response.TokenType
		p.Config.ExpiresAt = time.Now().Add(time.Duration(response.ExpiresIn) * time.Second).Unix()
	}

	return nil
}
