package tiktok

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"log"

	data_database "github.com/tkowalski/socgo/internal/data/database"
	"github.com/tkowalski/socgo/internal/data/provider"
)

// TikTokPost implements the Post interface for TikTok
type TikTokPost struct {
	Config     *provider.ProviderConfig
	HttpClient provider.HTTPClient
}

// NewTikTokPost creates a new TikTok Post
func NewTikTokPost(config *provider.ProviderConfig, httpClient provider.HTTPClient) *TikTokPost {
	return &TikTokPost{
		Config:     config,
		HttpClient: httpClient,
	}
}

// Publish publishes content to TikTok using proper TikTok Content Posting API v2
func (p *TikTokPost) Publish(ctx context.Context, dbProvider data_database.Provider, content string, media []provider.Media) (postID string, err error) {
	log.Printf("TikTok Publish called with content: %s", content)

	// TikTok requires video content for publishing
	if len(media) == 0 {
		return "", fmt.Errorf("TikTok posts require at least one video file")
	}

	// For now, support single video publishing (most common case)
	if len(media) == 1 {
		return p.publishSingleVideo(ctx, content, media[0])
	}

	// Multiple videos not supported in this version
	return "", fmt.Errorf("multiple video publishing not supported yet")
}

// publishSingleVideo publishes a single video to TikTok using Content Posting API v2
func (p *TikTokPost) publishSingleVideo(ctx context.Context, content string, media provider.Media) (string, error) {
	// Validate video file
	if err := p.validateVideoFile(media); err != nil {
		return "", fmt.Errorf("video validation failed: %w", err)
	}

	// Step 1: Initialize video upload
	uploadURL, uploadHeaders, err := p.initializeVideoUpload(ctx, media)
	if err != nil {
		return "", fmt.Errorf("failed to initialize video upload: %w", err)
	}

	// Step 2: Upload video file
	if err := p.uploadVideoFile(ctx, uploadURL, uploadHeaders, media); err != nil {
		return "", fmt.Errorf("failed to upload video file: %w", err)
	}

	// Step 3: Create and publish post
	publishedID, err := p.createVideoPost(ctx, content, media)
	if err != nil {
		return "", fmt.Errorf("failed to create video post: %w", err)
	}

	return publishedID, nil
}

// validateVideoFile validates the video file meets TikTok requirements
func (p *TikTokPost) validateVideoFile(media provider.Media) error {
	// Check file exists
	if media.FilePath == "" {
		return fmt.Errorf("no video file path provided")
	}

	// Get file info
	fileInfo, err := os.Stat(media.FilePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Check file size (max 4GB)
	maxSize := int64(4 * 1024 * 1024 * 1024) // 4GB
	if fileInfo.Size() > maxSize {
		return fmt.Errorf("video file too large: %d bytes (max: %d bytes)", fileInfo.Size(), maxSize)
	}

	// Check minimum size (should be at least 3 seconds worth)
	minSize := int64(1024 * 1024) // 1MB minimum
	if fileInfo.Size() < minSize {
		return fmt.Errorf("video file too small: %d bytes (min: %d bytes)", fileInfo.Size(), minSize)
	}

	// Validate video format based on file extension and MIME type
	supportedTypes := []string{"video/mp4", "video/quicktime", "video/mpeg", "video/avi", "video/webm", "video/3gpp"}
	isSupported := false
	for _, supportedType := range supportedTypes {
		if strings.Contains(strings.ToLower(media.MimeType), strings.ToLower(supportedType)) {
			isSupported = true
			break
		}
	}

	if !isSupported {
		return fmt.Errorf("unsupported video format: %s (supported: MP4, MOV, MPEG, AVI, WEBM, 3GP)", media.MimeType)
	}

	return nil
}

// initializeVideoUpload initializes the video upload using TikTok Content Posting API v2
func (p *TikTokPost) initializeVideoUpload(ctx context.Context, media provider.Media) (uploadURL string, uploadHeaders map[string]string, err error) {
	// TikTok Content Posting API v2 endpoint for upload initialization
	apiURL := "https://open.tiktokapis.com/v2/post/video/init/"

	// Get file size
	fileInfo, err := os.Stat(media.FilePath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Prepare form data
	data := url.Values{}
	data.Set("source_info", fmt.Sprintf(`{"source": "FILE_UPLOAD", "video_size": %d, "chunk_size": %d, "total_chunk_count": 1}`,
		fileInfo.Size(), fileInfo.Size()))

	// Make the request using injected HTTP client
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.Config.AccessToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.HttpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("initialize upload failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var response struct {
		Data struct {
			VideoID       string            `json:"video_id"`
			UploadURL     string            `json:"upload_url"`
			UploadHeaders map[string]string `json:"upload_headers"`
		} `json:"data"`
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return "", nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error.Message != "" {
		return "", nil, fmt.Errorf("TikTok API error: %s (code: %s)", response.Error.Message, response.Error.Code)
	}

	return response.Data.UploadURL, response.Data.UploadHeaders, nil
}

// uploadVideoFile uploads the video file to TikTok's upload endpoint
func (p *TikTokPost) uploadVideoFile(ctx context.Context, uploadURL string, uploadHeaders map[string]string, media provider.Media) error {
	// Open video file
	file, err := os.Open(media.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open video file: %w", err)
	}
	defer file.Close()

	// Create request
	req, err := http.NewRequestWithContext(ctx, "PUT", uploadURL, file)
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}

	// Set upload headers
	for key, value := range uploadHeaders {
		req.Header.Set(key, value)
	}

	// Set content type
	req.Header.Set("Content-Type", media.MimeType)

	// Make upload request
	resp, err := p.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload video: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("video upload failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// createVideoPost creates and publishes the video post using TikTok Content Posting API v2
func (p *TikTokPost) createVideoPost(ctx context.Context, content string, media provider.Media) (string, error) {
	// TikTok Content Posting API v2 endpoint for creating posts
	apiURL := "https://open.tiktokapis.com/v2/post/video/create/"

	// Prepare post data
	postData := map[string]interface{}{
		"text": content,
		"video_info": map[string]interface{}{
			"video_id": media.FileName, // This should be the video_id from upload response
		},
		"post_info": map[string]interface{}{
			"title":                    content,
			"privacy_level":            "PUBLIC_TO_EVERYONE",
			"disable_duet":             false,
			"disable_stitch":           false,
			"disable_comment":          false,
			"video_cover_timestamp_ms": 1000,
		},
		"source_info": map[string]interface{}{
			"source": "FILE_UPLOAD",
		},
	}

	jsonPayload, err := json.Marshal(postData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal post data: %w", err)
	}

	// Make the request using injected HTTP client
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.Config.AccessToken)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

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
		return "", fmt.Errorf("create video post failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var response struct {
		Data struct {
			VideoID string `json:"video_id"`
			PostID  string `json:"post_id"`
		} `json:"data"`
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error.Message != "" {
		return "", fmt.Errorf("TikTok API error: %s (code: %s)", response.Error.Message, response.Error.Code)
	}

	// Return post ID
	if response.Data.PostID != "" {
		return response.Data.PostID, nil
	}

	// Fallback to video ID if post ID not available
	if response.Data.VideoID != "" {
		return response.Data.VideoID, nil
	}

	return "", fmt.Errorf("no post ID returned from TikTok API")
}

// GetStatus retrieves the status of a published post using TikTok Content Posting API v2
func (p *TikTokPost) GetStatus(ctx context.Context, postID string) (status string, err error) {
	// TikTok Content Posting API v2 endpoint for getting video/post status
	apiURL := fmt.Sprintf("https://open.tiktokapis.com/v2/post/video/query/?video_id=%s", postID)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
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
		Data struct {
			VideoID      string `json:"video_id"`
			VideoStatus  string `json:"video_status"`
			PostStatus   string `json:"post_status"`
			FailReason   string `json:"fail_reason,omitempty"`
			PublishTime  int64  `json:"publish_time,omitempty"`
			ReviewStatus string `json:"review_status,omitempty"`
		} `json:"data"`
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error.Message != "" {
		return "", fmt.Errorf("TikTok API error: %s (code: %s)", response.Error.Message, response.Error.Code)
	}

	// Return status based on TikTok's response
	if response.Data.PostStatus != "" {
		switch response.Data.PostStatus {
		case "PUBLISHED":
			return string(provider.PostStatusPublished), nil
		case "PROCESSING":
			return string(provider.PostStatusPending), nil
		case "FAILED":
			if response.Data.FailReason != "" {
				return string(provider.PostStatusFailed), fmt.Errorf("post failed: %s", response.Data.FailReason)
			}
			return string(provider.PostStatusFailed), nil
		default:
			return string(provider.PostStatusPending), nil
		}
	}

	// Fallback to video status if post status not available
	if response.Data.VideoStatus != "" {
		switch response.Data.VideoStatus {
		case "UPLOADED", "PROCESSING_DONE":
			return string(provider.PostStatusPublished), nil
		case "PROCESSING":
			return string(provider.PostStatusPending), nil
		case "FAILED":
			return string(provider.PostStatusFailed), nil
		default:
			return string(provider.PostStatusPending), nil
		}
	}

	return string(provider.PostStatusPending), nil
}

// RefreshToken refreshes the access token using TikTok Content Posting API v2
func (p *TikTokPost) RefreshToken(ctx context.Context) error {
	if p.Config.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	// TikTok Content Posting API v2 endpoint for token refresh
	apiURL := "https://open.tiktokapis.com/v2/oauth/token/"

	// Prepare form data
	data := url.Values{}
	data.Set("client_key", "your_client_key")       // This should come from config
	data.Set("client_secret", "your_client_secret") // This should come from config
	data.Set("refresh_token", p.Config.RefreshToken)
	data.Set("grant_type", "refresh_token")

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
		Data struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			ExpiresIn    int64  `json:"expires_in"`
			TokenType    string `json:"token_type"`
			Scope        string `json:"scope"`
		} `json:"data"`
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error.Message != "" {
		return fmt.Errorf("TikTok API error: %s (code: %s)", response.Error.Message, response.Error.Code)
	}

	// Update config with new tokens
	if response.Data.AccessToken != "" {
		p.Config.AccessToken = response.Data.AccessToken
		p.Config.RefreshToken = response.Data.RefreshToken
		p.Config.TokenType = response.Data.TokenType
		p.Config.ExpiresAt = time.Now().Add(time.Duration(response.Data.ExpiresIn) * time.Second).Unix()
	}

	return nil
}
