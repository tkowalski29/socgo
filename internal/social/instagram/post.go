package instagram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"log"

	data_database "github.com/tkowalski/socgo/internal/data/database"
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

// validateInstagramConfig validates Instagram configuration before publishing
func (p *InstagramPost) validateInstagramConfig() error {
	if p.Config == nil {
		return fmt.Errorf("Instagram configuration is nil")
	}

	if p.Config.UserID == "" {
		return fmt.Errorf("Instagram UserID is required")
	}

	if p.Config.AccessToken == "" {
		return fmt.Errorf("Instagram AccessToken is required")
	}

	// Validate that UserID looks like a valid Instagram Business Account ID
	// Instagram Business Account IDs are typically numeric and 15-16 digits
	if !isValidInstagramBusinessAccountID(p.Config.UserID) {
		return fmt.Errorf("Invalid Instagram Business Account ID: %s. Please ensure you're using an Instagram Business Account connected to a Facebook Page", p.Config.UserID)
	}

	return nil
}

// isValidInstagramBusinessAccountID checks if the UserID looks like a valid Instagram Business Account ID
func isValidInstagramBusinessAccountID(userID string) bool {
	// Instagram Business Account IDs are typically numeric and 10-20 digits
	// Allow for different ID formats from Facebook Graph API
	if len(userID) < 8 || len(userID) > 25 {
		return false
	}

	// Check if it's numeric (Instagram Business Account IDs are always numeric)
	for _, char := range userID {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

// checkInstagramPermissions checks if the Instagram account has required permissions
func (p *InstagramPost) checkInstagramPermissions(ctx context.Context) error {
	log.Printf("🔍 Checking Instagram permissions for UserID: %s", p.Config.UserID)

	// First, try to get account info with business fields
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s?fields=id,username,account_type,media_count&access_token=%s",
		p.Config.UserID, p.Config.AccessToken)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create permissions check request: %w", err)
	}

	resp, err := p.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to check Instagram permissions: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read permissions response: %w", err)
	}

	log.Printf("Instagram permissions check response: %s", string(bodyBytes))

	// If we get an error about account_type field, try with basic fields
	if resp.StatusCode == http.StatusBadRequest && strings.Contains(string(bodyBytes), "account_type") {
		log.Printf("Account type field not available, trying basic account check")
		return p.checkBasicInstagramAccount(ctx)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Instagram permissions check failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		ID          string `json:"id"`
		Username    string `json:"username"`
		AccountType string `json:"account_type"`
		MediaCount  int    `json:"media_count"`
		Error       struct {
			Message   string `json:"message"`
			Type      string `json:"type"`
			Code      int    `json:"code"`
			ErrorCode int    `json:"error_code"`
		} `json:"error"`
	}

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return fmt.Errorf("failed to decode permissions response: %w", err)
	}

	if response.Error.Message != "" {
		return fmt.Errorf("Instagram permissions check error: %s (type: %s, code: %d)",
			response.Error.Message, response.Error.Type, response.Error.Code)
	}

	log.Printf("Instagram account info - ID: %s, Username: %s, AccountType: %s, MediaCount: %d",
		response.ID, response.Username, response.AccountType, response.MediaCount)

	// Check if it's a business account
	if response.AccountType != "BUSINESS" && response.AccountType != "CREATOR" {
		return fmt.Errorf(`Instagram account must be a Business or Creator account to publish content. 

Current account type: %s
Account: %s (@%s)

To fix this issue, you need to:

1. CONVERT YOUR INSTAGRAM ACCOUNT:
   - Open Instagram app or website
   - Go to Settings > Account
   - Select "Switch to Business Account"
   - Choose "Business Account" (not Creator)
   - Complete the conversion process

2. CREATE A FACEBOOK PAGE (if you don't have one):
   - Go to https://www.facebook.com/pages
   - Click "Create Page"
   - Choose "Business or Brand"
   - Enter page name (e.g., "My Business")
   - Choose category (e.g., "Software")
   - Complete page creation

3. CONNECT INSTAGRAM TO FACEBOOK PAGE:
   - Go to your Facebook Page
   - Settings > Instagram
   - Click "Connect Account"
   - Log in with your Instagram account

4. RECONNECT IN THIS APP:
   - Remove current Instagram connection
   - Connect Instagram again
   - The app should now find your Instagram Business Account`, response.AccountType, response.ID, response.Username)
	}

	log.Printf("✅ Instagram account verified as %s account - ready for publishing", response.AccountType)
	return nil
}

// checkBasicInstagramAccount checks basic Instagram account (personal account)
func (p *InstagramPost) checkBasicInstagramAccount(ctx context.Context) error {
	// For basic Instagram accounts, we can only read data, not publish
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s?fields=id,username&access_token=%s",
		p.Config.UserID, p.Config.AccessToken)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create basic account check request: %w", err)
	}

	resp, err := p.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to check basic Instagram account: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read basic account response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Basic Instagram account check failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Error    struct {
			Message   string `json:"message"`
			Type      string `json:"type"`
			Code      int    `json:"code"`
			ErrorCode int    `json:"error_code"`
		} `json:"error"`
	}

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return fmt.Errorf("failed to decode basic account response: %w", err)
	}

	if response.Error.Message != "" {
		return fmt.Errorf("Basic Instagram account check error: %s (type: %s, code: %d)",
			response.Error.Message, response.Error.Type, response.Error.Code)
	}

	log.Printf("Basic Instagram account info - ID: %s, Username: %s", response.ID, response.Username)

	// For basic Instagram accounts, we cannot publish content
	return fmt.Errorf(`Instagram personal accounts cannot publish content via API. 

To fix this issue, you need to:

1. CONVERT YOUR INSTAGRAM ACCOUNT:
   - Open Instagram app or website
   - Go to Settings > Account
   - Select "Switch to Business Account"
   - Choose "Business Account" (not Creator)
   - Complete the conversion process

2. CREATE A FACEBOOK PAGE (if you don't have one):
   - Go to https://www.facebook.com/pages
   - Click "Create Page"
   - Choose "Business or Brand"
   - Enter page name (e.g., "My Business")
   - Choose category (e.g., "Software")
   - Complete page creation

3. CONNECT INSTAGRAM TO FACEBOOK PAGE:
   - Go to your Facebook Page
   - Settings > Instagram
   - Click "Connect Account"
   - Log in with your Instagram account

4. RECONNECT IN THIS APP:
   - Remove current Instagram connection
   - Connect Instagram again
   - The app should now find your Instagram Business Account

Current account: %s (@%s)`, response.ID, response.Username)
}

// Publish publishes content to Instagram using proper Instagram Graph API
func (p *InstagramPost) Publish(ctx context.Context, dbProvider data_database.Provider, content string, media []provider.Media) (postID string, err error) {
	log.Printf("Instagram Publish called with content: %s", content)
	log.Printf("🔧 Instagram configuration:")
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
	if err := p.validateInstagramConfig(); err != nil {
		return "", fmt.Errorf("Instagram configuration validation failed: %w", err)
	}

	// Check permissions
	if err := p.checkInstagramPermissions(ctx); err != nil {
		return "", fmt.Errorf("Instagram permissions check failed: %w", err)
	}

	// Instagram requires 2-step publishing process:
	// 1. Create media container(s)
	// 2. Publish the container(s)

	if len(media) > 0 {
		// Publish with media using Instagram Graph API
		return p.publishWithMedia(ctx, content, media)
	}

	// For text-only posts, Instagram doesn't support them directly
	// Instagram requires at least one media item for posts
	return "", fmt.Errorf("Instagram posts require at least one media item (image or video)")
}

// publishWithMedia publishes Instagram post with media using proper Instagram Graph API
func (p *InstagramPost) publishWithMedia(ctx context.Context, content string, media []provider.Media) (string, error) {
	if len(media) == 0 {
		return "", fmt.Errorf("no media provided")
	}

	// For single media, use single media publishing
	if len(media) == 1 {
		return p.publishSingleMedia(ctx, content, media[0])
	}

	// For multiple media, use carousel publishing
	return p.publishCarousel(ctx, content, media)
}

// publishSingleMedia publishes a single image/video to Instagram
func (p *InstagramPost) publishSingleMedia(ctx context.Context, content string, media provider.Media) (string, error) {
	// Step 1: Create media container
	containerID, err := p.createMediaContainer(ctx, content, media, "")
	if err != nil {
		return "", fmt.Errorf("failed to create media container: %w", err)
	}

	// Step 2: Publish the container
	publishedID, err := p.publishContainer(ctx, containerID)
	if err != nil {
		return "", fmt.Errorf("failed to publish media container: %w", err)
	}

	return publishedID, nil
}

// publishCarousel publishes multiple images as a carousel to Instagram
func (p *InstagramPost) publishCarousel(ctx context.Context, content string, media []provider.Media) (string, error) {
	var childContainers []string

	// Step 1: Create child containers for each media item
	for _, m := range media {
		containerID, err := p.createMediaContainer(ctx, "", m, "CAROUSEL_ALBUM")
		if err != nil {
			return "", fmt.Errorf("failed to create child container for %s: %w", m.FileName, err)
		}
		childContainers = append(childContainers, containerID)
	}

	// Step 2: Create parent carousel container
	carouselContainerID, err := p.createCarouselContainer(ctx, content, childContainers)
	if err != nil {
		return "", fmt.Errorf("failed to create carousel container: %w", err)
	}

	// Step 3: Publish the carousel
	publishedID, err := p.publishContainer(ctx, carouselContainerID)
	if err != nil {
		return "", fmt.Errorf("failed to publish carousel: %w", err)
	}

	return publishedID, nil
}

// createMediaContainer creates a media container on Instagram
func (p *InstagramPost) createMediaContainer(ctx context.Context, caption string, media provider.Media, mediaType string) (string, error) {
	// Instagram Graph API endpoint for creating media containers
	apiURL := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/media", p.Config.UserID)

	log.Printf("Creating media container for Instagram account: %s", p.Config.UserID)
	log.Printf("API URL: %s", apiURL)
	log.Printf("Media type: %s, File: %s", mediaType, media.FileName)

	// Determine media type if not specified
	if mediaType == "" {
		mediaType = "IMAGE"
		if strings.Contains(strings.ToLower(media.MimeType), "video") {
			mediaType = "VIDEO"
		}
	}

	// Prepare form data
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	// Add access token
	if err := writer.WriteField("access_token", p.Config.AccessToken); err != nil {
		return "", fmt.Errorf("failed to write access_token field: %w", err)
	}

	// Add media type
	if err := writer.WriteField("media_type", mediaType); err != nil {
		return "", fmt.Errorf("failed to write media_type field: %w", err)
	}

	// Add caption if provided
	if caption != "" {
		if err := writer.WriteField("caption", caption); err != nil {
			return "", fmt.Errorf("failed to write caption field: %w", err)
		}
	}

	// Upload file directly
	if media.FilePath != "" {
		// Upload file directly
		file, err := os.Open(media.FilePath)
		if err != nil {
			return "", fmt.Errorf("failed to open file %s: %w", media.FilePath, err)
		}
		defer file.Close()

		fieldName := "source"
		if mediaType == "VIDEO" {
			fieldName = "video_url" // Instagram requires video_url for videos
		}

		part, err := writer.CreateFormFile(fieldName, media.FileName)
		if err != nil {
			return "", fmt.Errorf("failed to create form file: %w", err)
		}

		if _, err := io.Copy(part, file); err != nil {
			return "", fmt.Errorf("failed to copy file data: %w", err)
		}
	} else {
		return "", fmt.Errorf("no file path provided")
	}

	writer.Close()

	log.Printf("Making request to Instagram Graph API...")

	// Make the request using injected HTTP client
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, &b)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

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

	log.Printf("Instagram API response status: %d", resp.StatusCode)
	log.Printf("Instagram API response body: %s", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("create container failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
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

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error.Message != "" {
		return "", fmt.Errorf("Instagram API error: %s (type: %s, code: %d)",
			response.Error.Message, response.Error.Type, response.Error.Code)
	}

	log.Printf("Successfully created media container with ID: %s", response.ID)
	return response.ID, nil
}

// createCarouselContainer creates a carousel container for multiple media
func (p *InstagramPost) createCarouselContainer(ctx context.Context, caption string, childContainers []string) (string, error) {
	// Instagram Graph API endpoint for creating carousel containers
	apiURL := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/media", p.Config.UserID)

	// Prepare form data
	data := url.Values{}
	data.Set("access_token", p.Config.AccessToken)
	data.Set("media_type", "CAROUSEL_ALBUM")
	data.Set("caption", caption)

	// Add child containers
	for i, containerID := range childContainers {
		data.Set(fmt.Sprintf("children[%d]", i), containerID)
	}

	// Make the request using injected HTTP client
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("create carousel container failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
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

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error.Message != "" {
		return "", fmt.Errorf("Instagram API error: %s (type: %s, code: %d)",
			response.Error.Message, response.Error.Type, response.Error.Code)
	}

	return response.ID, nil
}

// publishContainer publishes a created media container
func (p *InstagramPost) publishContainer(ctx context.Context, containerID string) (string, error) {
	// Instagram Graph API endpoint for publishing containers
	apiURL := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/media_publish", p.Config.UserID)

	// Prepare form data
	data := url.Values{}
	data.Set("access_token", p.Config.AccessToken)
	data.Set("creation_id", containerID)

	// Make the request using injected HTTP client
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("publish container failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
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

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error.Message != "" {
		return "", fmt.Errorf("Instagram API error: %s (type: %s, code: %d)",
			response.Error.Message, response.Error.Type, response.Error.Code)
	}

	return response.ID, nil
}

// GetStatus retrieves the status of a published post using Instagram Graph API
func (p *InstagramPost) GetStatus(ctx context.Context, postID string) (status string, err error) {
	// Instagram Graph API endpoint for getting post status
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s?fields=id,media_type,permalink,timestamp,media_url&access_token=%s",
		postID, p.Config.AccessToken)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

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
		ID        string `json:"id"`
		MediaType string `json:"media_type"`
		Permalink string `json:"permalink"`
		Timestamp string `json:"timestamp"`
		MediaURL  string `json:"media_url"`
		Error     struct {
			Message   string `json:"message"`
			Type      string `json:"type"`
			Code      int    `json:"code"`
			ErrorCode int    `json:"error_code"`
		} `json:"error"`
	}

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error.Message != "" {
		return "", fmt.Errorf("Instagram API error: %s (type: %s, code: %d)",
			response.Error.Message, response.Error.Type, response.Error.Code)
	}

	// Return status based on response
	if response.ID != "" && response.Timestamp != "" {
		return string(provider.PostStatusPublished), nil
	}

	return string(provider.PostStatusPending), nil
}

// RefreshToken refreshes the access token using Instagram Graph API
func (p *InstagramPost) RefreshToken(ctx context.Context) error {
	if p.Config.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	// Instagram uses Facebook's token refresh mechanism
	// Since Instagram Business accounts use Facebook page tokens,
	// we use Facebook's token exchange endpoint
	apiURL := "https://graph.facebook.com/v18.0/oauth/access_token"

	// Prepare form data
	data := url.Values{}
	data.Set("grant_type", "fb_exchange_token")
	data.Set("client_id", "your_app_id")         // This should come from config
	data.Set("client_secret", "your_app_secret") // This should come from config
	data.Set("fb_exchange_token", p.Config.AccessToken)

	// Make the request using injected HTTP client
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
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
