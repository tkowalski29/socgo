package facebook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"log"

	"github.com/tkowalski/socgo/internal/data/provider"
)

// FacebookPost implements the Post interface for Facebook
type FacebookPost struct {
	Config     *provider.ProviderConfig
	HttpClient provider.HTTPClient
}

// NewFacebookPost creates a new Facebook Post
func NewFacebookPost(config *provider.ProviderConfig, httpClient provider.HTTPClient) *FacebookPost {
	return &FacebookPost{
		Config:     config,
		HttpClient: httpClient,
	}
}

// Funkcja pomocnicza do doklejania lokalizacji
func appendLocation(content string, settings map[string]string) string {
	if settings != nil {
		if location, ok := settings["facebook_location"]; ok && location != "" {
			log.Printf("Adding location to content: %s", location)
			return content + "\n📍 " + location
		}
	}
	log.Printf("No location found in settings or settings is nil")
	return content
}

// Publish publishes content to Facebook
func (p *FacebookPost) Publish(ctx context.Context, content string, media []provider.Media) (postID string, err error) {
	log.Printf("Publish called with original content: %s", content)
	log.Printf("Settings: %+v", p.Config.Settings)

	content = appendLocation(content, p.Config.Settings)
	log.Printf("Content after adding location: %s", content)

	// If we have media files, try to publish with media
	if len(media) > 0 {
		// Use publishMultipleImages for all media cases (it handles single and multiple images)
		postID, err = p.publishMultipleImages(ctx, content, media)
		if err != nil {
			// Check if it's a duplicate media error
			if strings.Contains(err.Error(), "Already Posted") || strings.Contains(err.Error(), "error_subcode\":1366051") {
				log.Printf("Warning: Media already posted, publishing text-only post")
				// Fallback to text-only post
				return p.publishTextOnly(ctx, content)
			} else {
				log.Printf("Warning: Failed to publish with media: %v, falling back to text-only post", err)
				// Fallback to text-only post
				return p.publishTextOnly(ctx, content)
			}
		}
		return postID, nil
	}

	// If no media, publish text-only post
	return p.publishTextOnly(ctx, content)
}

// publishTextOnly publishes a text-only post to Facebook
func (p *FacebookPost) publishTextOnly(ctx context.Context, content string) (string, error) {
	log.Printf("publishTextOnly called with content: %s", content)

	// Facebook Graph API endpoint for publishing to page
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/feed", p.Config.UserID)

	// Prepare request payload
	payload := map[string]interface{}{
		"message":      content,
		"access_token": p.Config.AccessToken,
	}

	log.Printf("Sending payload to Facebook: %+v", payload)

	// Marshal payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.HttpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("text-only API request failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
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

	if response.ID != "" {
		return response.ID, nil
	}

	return fmt.Sprintf("facebook_%d", time.Now().UnixNano()), nil
}

// GetStatus retrieves the status of a published post
func (p *FacebookPost) GetStatus(ctx context.Context, postID string) (status string, err error) {
	// Facebook Graph API endpoint for getting post status
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s?fields=id,message,created_time,status_type&access_token=%s",
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

	if resp.StatusCode != http.StatusOK {
		// Read response body for better error information
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
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

	// Return status based on response
	if response.ID != "" && response.CreatedTime != "" {
		return string(provider.PostStatusPublished), nil
	}

	return string(provider.PostStatusPending), nil
}

// RefreshToken refreshes the access token
func (p *FacebookPost) RefreshToken(ctx context.Context) error {
	if p.Config.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	// Facebook Graph API endpoint for token refresh (mock implementation)
	url := "https://graph.facebook.com/oauth/access_token"

	payload := map[string]interface{}{
		"grant_type":        "fb_exchange_token",
		"client_id":         "your_app_id",
		"client_secret":     "your_app_secret",
		"fb_exchange_token": p.Config.AccessToken,
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
		p.Config.AccessToken = response.AccessToken
		p.Config.TokenType = response.TokenType
		p.Config.ExpiresAt = time.Now().Add(time.Duration(response.ExpiresIn) * time.Second).Unix()
	}

	return nil
}

// publishWithMedia publishes a post with attached media directly to the feed
func (p *FacebookPost) publishWithMedia(ctx context.Context, content string, media []provider.Media) (string, error) {
	if len(media) == 0 {
		return "", fmt.Errorf("no media provided")
	}

	// Facebook Graph API endpoint for posting photo with message
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/photos", p.Config.UserID)

	// Open the file
	file, err := os.Open(media[0].FilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", media[0].FilePath, err)
	}
	defer file.Close()

	// Create multipart form data
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	// Add access token
	if err := writer.WriteField("access_token", p.Config.AccessToken); err != nil {
		return "", fmt.Errorf("failed to write access_token field: %w", err)
	}

	// Add message
	if content != "" {
		if err := writer.WriteField("message", content); err != nil {
			return "", fmt.Errorf("failed to write message field: %w", err)
		}
	}

	// Add file as source
	part, err := writer.CreateFormFile("source", media[0].FileName)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return "", fmt.Errorf("failed to copy file data: %w", err)
	}

	// Close the multipart writer
	writer.Close()

	// Make the request
	resp, err := http.Post(url, writer.FormDataContentType(), &b)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("media post failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response from bodyBytes instead of resp.Body
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
		return "", fmt.Errorf("Facebook API error: %s (type: %s, code: %d)",
			response.Error.Message, response.Error.Type, response.Error.Code)
	}

	return response.ID, nil
}

// publishMultipleImages publishes a post with multiple images to Facebook
func (p *FacebookPost) publishMultipleImages(ctx context.Context, content string, media []provider.Media) (string, error) {
	if len(media) < 1 {
		return "", fmt.Errorf("publishMultipleImages wymaga co najmniej 1 obrazek")
	}

	// For single image, use simpler approach
	if len(media) == 1 {
		return p.publishWithMedia(ctx, content, media)
	}

	var mediaIDs []string
	for _, m := range media {
		// 1. Upload each image as unpublished
		id, err := p.uploadPhotoUnpublished(ctx, m)
		if err != nil {
			return "", fmt.Errorf("failed to upload image: %w", err)
		}
		mediaIDs = append(mediaIDs, id)
	}

	// 2. Create post with attached_media
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/feed", p.Config.UserID)

	// Prepare multipart form data
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	// Add access token
	if err := writer.WriteField("access_token", p.Config.AccessToken); err != nil {
		return "", fmt.Errorf("failed to write access_token field: %w", err)
	}

	// Add message
	if content != "" {
		if err := writer.WriteField("message", content); err != nil {
			return "", fmt.Errorf("failed to write message field: %w", err)
		}
	}

	// Add attached_media fields
	for i, id := range mediaIDs {
		field := fmt.Sprintf("attached_media[%d]", i)
		jsonVal := fmt.Sprintf(`{"media_fbid":"%s"}`, id)
		if err := writer.WriteField(field, jsonVal); err != nil {
			return "", fmt.Errorf("failed to write attached_media field: %w", err)
		}
	}

	writer.Close()

	resp, err := http.Post(url, writer.FormDataContentType(), &b)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("multi-image post failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

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
		return "", fmt.Errorf("Facebook API error: %s (type: %s, code: %d)",
			response.Error.Message, response.Error.Type, response.Error.Code)
	}
	return response.ID, nil
}

// uploadPhotoUnpublished uploads a photo to Facebook as unpublished and returns the media_fbid
func (p *FacebookPost) uploadPhotoUnpublished(ctx context.Context, m provider.Media) (string, error) {
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/photos", p.Config.UserID)

	file, err := os.Open(m.FilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", m.FilePath, err)
	}
	defer file.Close()

	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	if err := writer.WriteField("access_token", p.Config.AccessToken); err != nil {
		return "", fmt.Errorf("failed to write access_token field: %w", err)
	}
	if err := writer.WriteField("published", "false"); err != nil {
		return "", fmt.Errorf("failed to write published field: %w", err)
	}
	part, err := writer.CreateFormFile("source", m.FileName)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", fmt.Errorf("failed to copy file data: %w", err)
	}
	writer.Close()

	resp, err := http.Post(url, writer.FormDataContentType(), &b)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("upload unpublished failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}
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
		return "", fmt.Errorf("Facebook API error: %s (type: %s, code: %d)",
			response.Error.Message, response.Error.Type, response.Error.Code)
	}
	return response.ID, nil
}
