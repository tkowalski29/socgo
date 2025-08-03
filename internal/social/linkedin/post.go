package linkedin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
		"author":         fmt.Sprintf("urn:li:person:%s", p.Config.UserID),
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

// publishWithMedia publishes a post with media to LinkedIn using 3-step process
func (p *LinkedInPost) publishWithMedia(ctx context.Context, content string, media []provider.Media) (string, error) {
	// LinkedIn supports multiple images (up to 9) in UGC Posts API
	if len(media) == 0 {
		return "", fmt.Errorf("no media provided")
	}

	log.Printf("📸 Publishing LinkedIn post with %d media files", len(media))

	// LinkedIn supports up to 9 images
	if len(media) > 9 {
		log.Printf("⚠️ LinkedIn supports max 9 images, using first 9 of %d provided", len(media))
		media = media[:9]
	}

	// Upload all media files and collect assets
	var assets []string

	for i, mediaItem := range media {
		log.Printf("📸 Processing media %d/%d: %s (type: %s)", i+1, len(media), mediaItem.FileName, mediaItem.MimeType)

		// Check if media type is supported
		if !p.isMediaTypeSupported(mediaItem.MimeType) {
			log.Printf("⚠️ Skipping unsupported media type: %s", mediaItem.MimeType)
			continue
		}

		// LinkedIn 3-step process for each media file:
		// 1. Register upload
		asset, uploadURL, err := p.registerMediaUpload(ctx)
		if err != nil {
			log.Printf("❌ Failed to register media upload for %s: %v", mediaItem.FileName, err)
			continue
		}

		// 2. Upload binary data
		err = p.uploadMediaBinary(ctx, uploadURL, mediaItem)
		if err != nil {
			log.Printf("❌ Failed to upload media binary for %s: %v", mediaItem.FileName, err)
			continue
		}

		assets = append(assets, asset)
		log.Printf("✅ Successfully uploaded media %d/%d: %s", i+1, len(media), mediaItem.FileName)
	}

	// Check if we have any successfully uploaded assets
	if len(assets) == 0 {
		log.Printf("❌ No media files were successfully uploaded. Publishing as text-only")
		var mediaNames []string
		for _, m := range media {
			mediaNames = append(mediaNames, m.FileName)
		}
		contentWithMedia := content + fmt.Sprintf("\n\n[Media files: %v]", mediaNames)
		return p.publishTextOnly(ctx, contentWithMedia)
	}

	log.Printf("✅ Successfully uploaded %d/%d media files", len(assets), len(media))

	// 3. Create UGC post with all media references
	return p.publishWithLinkedInMediaMultiple(ctx, content, assets)
}

// isMediaTypeSupported checks if the media type is supported by LinkedIn
func (p *LinkedInPost) isMediaTypeSupported(mimeType string) bool {
	supportedTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
	}

	for _, supportedType := range supportedTypes {
		if mimeType == supportedType {
			return true
		}
	}
	return false
}

// registerMediaUpload registers an upload with LinkedIn and returns upload token and URL
func (p *LinkedInPost) registerMediaUpload(ctx context.Context) (string, string, error) {
	apiURL := "https://api.linkedin.com/v2/assets?action=registerUpload"

	payload := map[string]interface{}{
		"registerUploadRequest": map[string]interface{}{
			"recipes": []string{"urn:li:digitalmediaRecipe:feedshare-image"},
			"owner":   fmt.Sprintf("urn:li:person:%s", p.Config.UserID),
			"serviceRelationships": []map[string]interface{}{
				{
					"relationshipType": "OWNER",
					"identifier":       "urn:li:userGeneratedContent",
				},
			},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal register upload payload: %w", err)
	}

	log.Printf("🔧 Registering LinkedIn media upload...")

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.Config.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Restli-Protocol-Version", "2.0.0")

	resp, err := p.HttpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read response body: %w", err)
	}

	log.Printf("📤 Register upload response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("register upload failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		Value struct {
			UploadMechanism struct {
				ComLinkedinDigitalmediaUploadingMediaUploadHttpRequest struct {
					UploadURL string            `json:"uploadUrl"`
					Headers   map[string]string `json:"headers"`
				} `json:"com.linkedin.digitalmedia.uploading.MediaUploadHttpRequest"`
			} `json:"uploadMechanism"`
			Asset string `json:"asset"`
		} `json:"value"`
	}

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return "", "", fmt.Errorf("failed to decode response: %w", err)
	}

	uploadURL := response.Value.UploadMechanism.ComLinkedinDigitalmediaUploadingMediaUploadHttpRequest.UploadURL
	asset := response.Value.Asset

	if uploadURL == "" || asset == "" {
		return "", "", fmt.Errorf("invalid response: missing uploadUrl or asset")
	}

	log.Printf("✅ Upload registered. Asset: %s", asset)
	return asset, uploadURL, nil
}

// uploadMediaBinary uploads the actual binary data to LinkedIn
func (p *LinkedInPost) uploadMediaBinary(ctx context.Context, uploadURL string, media provider.Media) error {
	log.Printf("📤 Uploading media binary to LinkedIn...")

	// Read file data from FilePath
	file, err := os.Open(media.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open media file %s: %w", media.FilePath, err)
	}
	defer file.Close()

	// Read file content
	fileData, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read media file %s: %w", media.FilePath, err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", uploadURL, bytes.NewReader(fileData))
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := p.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload media: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("📤 Media upload response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("media upload failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	log.Printf("✅ Media uploaded successfully")
	return nil
}

// publishWithLinkedInMediaMultiple creates a UGC post with multiple media references
func (p *LinkedInPost) publishWithLinkedInMediaMultiple(ctx context.Context, content string, assets []string) (string, error) {
	apiURL := "https://api.linkedin.com/v2/ugcPosts"

	// Build media array for LinkedIn UGC API
	var mediaArray []map[string]interface{}

	for i, asset := range assets {
		mediaArray = append(mediaArray, map[string]interface{}{
			"status": "READY",
			"description": map[string]interface{}{
				"text": fmt.Sprintf("Image %d", i+1),
			},
			"media": asset,
			"title": map[string]interface{}{
				"text": fmt.Sprintf("Image %d", i+1),
			},
		})
	}

	payload := map[string]interface{}{
		"author":         fmt.Sprintf("urn:li:person:%s", p.Config.UserID),
		"lifecycleState": "PUBLISHED",
		"specificContent": map[string]interface{}{
			"com.linkedin.ugc.ShareContent": map[string]interface{}{
				"shareCommentary": map[string]interface{}{
					"text": content,
				},
				"shareMediaCategory": "IMAGE",
				"media":              mediaArray,
			},
		},
		"visibility": map[string]interface{}{
			"com.linkedin.ugc.MemberNetworkVisibility": "PUBLIC",
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal LinkedIn post with multiple media payload: %w", err)
	}

	log.Printf("🚀 Creating LinkedIn post with %d media files...", len(assets))
	log.Printf("📝 Payload: %s", string(jsonPayload))

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

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	log.Printf("🚀 LinkedIn post with multiple media response status: %d", resp.StatusCode)
	log.Printf("📋 Response body: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("post with multiple media failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		ID string `json:"id"`
	}

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("✅ Successfully published LinkedIn post with %d media files. ID: %s", len(assets), response.ID)
	return response.ID, nil
}

// publishWithLinkedInMedia creates a UGC post with single media reference (legacy for compatibility)
func (p *LinkedInPost) publishWithLinkedInMedia(ctx context.Context, content, asset string) (string, error) {
	apiURL := "https://api.linkedin.com/v2/ugcPosts"

	payload := map[string]interface{}{
		"author":         fmt.Sprintf("urn:li:person:%s", p.Config.UserID),
		"lifecycleState": "PUBLISHED",
		"specificContent": map[string]interface{}{
			"com.linkedin.ugc.ShareContent": map[string]interface{}{
				"shareCommentary": map[string]interface{}{
					"text": content,
				},
				"shareMediaCategory": "IMAGE",
				"media": []map[string]interface{}{
					{
						"status": "READY",
						"description": map[string]interface{}{
							"text": "Shared image",
						},
						"media": asset,
						"title": map[string]interface{}{
							"text": "Image",
						},
					},
				},
			},
		},
		"visibility": map[string]interface{}{
			"com.linkedin.ugc.MemberNetworkVisibility": "PUBLIC",
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal LinkedIn post with media payload: %w", err)
	}

	log.Printf("🚀 Creating LinkedIn post with media...")
	log.Printf("📝 Payload: %s", string(jsonPayload))

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

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	log.Printf("🚀 LinkedIn post with media response status: %d", resp.StatusCode)
	log.Printf("📋 Response body: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("post with media failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		ID string `json:"id"`
	}

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("✅ Successfully published LinkedIn post with media. ID: %s", response.ID)
	return response.ID, nil
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
