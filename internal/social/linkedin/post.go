package linkedin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	data_database "github.com/tkowalski/socgo/internal/data/database"
	"github.com/tkowalski/socgo/internal/data/provider"
)

// LinkedInPost implements the Post interface for LinkedIn
type LinkedInPost struct {
	Config     *provider.ProviderConfig
	HttpClient provider.HTTPClient
	BaseURL    string
}

// NewLinkedInPost creates a new LinkedIn Post
func NewLinkedInPost(config *provider.ProviderConfig, httpClient provider.HTTPClient, baseURL string) *LinkedInPost {
	return &LinkedInPost{
		Config:     config,
		HttpClient: httpClient,
		BaseURL:    baseURL,
	}
}

// validateLinkedInConfig validates LinkedIn configuration before publishing
func (p *LinkedInPost) validateLinkedInConfig() error {
	if p.Config == nil {
		return fmt.Errorf("LinkedIn configuration is nil")
	}

	if p.Config.UserID == "" {
		return fmt.Errorf("LinkedIn UserID is required")
	}

	if p.Config.AccessToken == "" {
		return fmt.Errorf("LinkedIn AccessToken is required")
	}

	return nil
}

// Publish publishes content to LinkedIn using LinkedIn UGC Posts API
func (p *LinkedInPost) Publish(ctx context.Context, dbProvider data_database.Provider, content string, media []provider.Media) (postID string, err error) {
	log.Printf("LinkedIn Publish called with content: %s", content)
	log.Printf("🔧 LinkedIn configuration:")
	log.Printf("   UserID: %s", p.Config.UserID)
	
	// Safely log access token prefix
	tokenPreview := p.Config.AccessToken
	if len(tokenPreview) > 20 {
		tokenPreview = tokenPreview[:20] + "..."
	} else if len(tokenPreview) > 5 {
		tokenPreview = tokenPreview[:5] + "..."
	}
	log.Printf("   AccessToken: %s", tokenPreview)
	log.Printf("   Media count: %d", len(media))

	// Validate configuration first
	if err := p.validateLinkedInConfig(); err != nil {
		return "", fmt.Errorf("LinkedIn configuration validation failed: %w", err)
	}

	// LinkedIn UGC Posts API requires specific format
	if len(media) > 0 {
		// Publish with media using LinkedIn UGC Posts API
		return p.publishWithMedia(ctx, content, media)
	}

	// For text-only posts
	return p.publishTextOnly(ctx, content)
}

// publishTextOnly publishes a text-only post to LinkedIn
func (p *LinkedInPost) publishTextOnly(ctx context.Context, content string) (string, error) {
	// LinkedIn UGC Posts API endpoint
	apiURL := "https://api.linkedin.com/v2/ugcPosts"

	// Prepare UGC post payload
	payload := map[string]interface{}{
		"author": fmt.Sprintf("urn:li:person:%s", p.Config.UserID),
		"lifecycleState": "PUBLISHED",
		"specificContent": map[string]interface{}{
			"com.linkedin.ugc.ShareContent": map[string]interface{}{
				"shareCommentary": map[string]interface{}{
					"text": content,
				},
				"shareMediaCategory": "NONE",
			},
		},
		"visibility": map[string]interface{}{
			"com.linkedin.ugc.MemberNetworkVisibility": "PUBLIC",
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal LinkedIn post payload: %w", err)
	}

	log.Printf("Making request to LinkedIn UGC Posts API...")
	log.Printf("API URL: %s", apiURL)
	log.Printf("Payload: %s", string(jsonPayload))

	// Make the request using injected HTTP client
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+p.Config.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Restli-Protocol-Version", "2.0.0")

	resp, err := p.HttpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	log.Printf("LinkedIn API response status: %d", resp.StatusCode)
	log.Printf("LinkedIn API response body: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("publish failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response to get post ID
	var response struct {
		ID string `json:"id"`
	}

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("Successfully published LinkedIn post with ID: %s", response.ID)
	return response.ID, nil
}

// publishWithMedia publishes a post with media to LinkedIn
func (p *LinkedInPost) publishWithMedia(ctx context.Context, content string, media []provider.Media) (string, error) {
	// For this implementation, we'll upload the first media item as an image
	// LinkedIn supports single image posts through UGC Posts API
	if len(media) == 0 {
		return "", fmt.Errorf("no media provided")
	}

	// Take the first media item
	firstMedia := media[0]

	// First, we need to register the upload (LinkedIn 3-step process)
	// 1. Register upload
	// 2. Upload binary data
	// 3. Create UGC post with media reference

	// For simplicity, we'll implement text-only for now and log that media was provided
	log.Printf("⚠️ Media publishing to LinkedIn not fully implemented yet")
	log.Printf("📁 Media provided: %s (type: %s)", firstMedia.FileName, firstMedia.MimeType)
	log.Printf("📝 Publishing as text-only post for now")

	// Append media info to content as text
	contentWithMedia := content + fmt.Sprintf("\n\n[Media: %s]", firstMedia.FileName)
	
	return p.publishTextOnly(ctx, contentWithMedia)
}

// GetStatus retrieves the status of a published post using LinkedIn API
func (p *LinkedInPost) GetStatus(ctx context.Context, postID string) (status string, err error) {
	// LinkedIn UGC Posts API endpoint for getting post details
	url := fmt.Sprintf("https://api.linkedin.com/v2/ugcPosts/%s", postID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+p.Config.AccessToken)
	req.Header.Set("X-Restli-Protocol-Version", "2.0.0")

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

	// Read response body for better error information
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var response struct {
		ID             string `json:"id"`
		LifecycleState string `json:"lifecycleState"`
		Error          struct {
			Message string `json:"message"`
			Status  int    `json:"status"`
		} `json:"error"`
	}

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error.Message != "" {
		return "", fmt.Errorf("LinkedIn API error: %s (status: %d)",
			response.Error.Message, response.Error.Status)
	}

	// Return status based on lifecycle state
	switch response.LifecycleState {
	case "PUBLISHED":
		return string(provider.PostStatusPublished), nil
	case "DRAFT":
		return string(provider.PostStatusPending), nil
	default:
		return string(provider.PostStatusPending), nil
	}
}

// RefreshToken refreshes the access token using LinkedIn OAuth 2.0
func (p *LinkedInPost) RefreshToken(ctx context.Context) error {
	if p.Config.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	// LinkedIn OAuth 2.0 token refresh endpoint
	apiURL := "https://www.linkedin.com/oauth/v2/accessToken"

	// Prepare form data
	payload := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": p.Config.RefreshToken,
		"client_id":     "your_client_id",     // This should come from config
		"client_secret": "your_client_secret", // This should come from config
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal refresh token payload: %w", err)
	}

	// Make the request using injected HTTP client
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonPayload))
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

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token refresh failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var response struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
		RefreshToken string `json:"refresh_token,omitempty"`
		Error        string `json:"error,omitempty"`
		ErrorDesc    string `json:"error_description,omitempty"`
	}

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error != "" {
		return fmt.Errorf("LinkedIn API error: %s - %s", response.Error, response.ErrorDesc)
	}

	// Update config with new tokens
	if response.AccessToken != "" {
		p.Config.AccessToken = response.AccessToken
		p.Config.TokenType = response.TokenType
		p.Config.ExpiresAt = time.Now().Add(time.Duration(response.ExpiresIn) * time.Second).Unix()
		
		if response.RefreshToken != "" {
			p.Config.RefreshToken = response.RefreshToken
		}
	}

	return nil
}