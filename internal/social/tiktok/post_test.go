package tiktok

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/tkowalski/socgo/internal/data/provider"
)

// MockHTTPClient is a mock implementation of HTTPClient for testing
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestTikTokPost_Publish(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		media          []provider.Media
		mockResponses  []string // Multiple responses for multi-step process
		mockStatusCodes []int   // Multiple status codes
		expectedPostID string
		expectError    bool
	}{
		{
			name:    "no media error",
			content: "Test content",
			media:   []provider.Media{},
			mockResponses: []string{},
			mockStatusCodes: []int{},
			expectedPostID: "",
			expectError:    true,
		},
		{
			name:    "multiple videos not supported",
			content: "Test content",
			media: []provider.Media{
				{FileName: "test1.mp4", FilePath: "/tmp/test1.mp4", MimeType: "video/mp4"},
				{FileName: "test2.mp4", FilePath: "/tmp/test2.mp4", MimeType: "video/mp4"},
			},
			mockResponses: []string{},
			mockStatusCodes: []int{},
			expectedPostID: "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock HTTP client
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// For simple error cases, just return without specific validation
					if len(tt.mockResponses) == 0 {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
						}, nil
					}

					// For multi-step processes, return appropriate response
					return &http.Response{
						StatusCode: tt.mockStatusCodes[0],
						Body:       io.NopCloser(bytes.NewBufferString(tt.mockResponses[0])),
					}, nil
				},
			}

			// Create Post
			config := &provider.ProviderConfig{
				AccessToken: "test_token",
				UserID:      "test_user",
			}
			Post := NewTikTokPost(config, mockClient)

			// Test Publish
			postID, err := Post.Publish(context.Background(), tt.content, tt.media)

			// Verify results
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if postID != tt.expectedPostID {
					t.Errorf("Expected postID %s, got %s", tt.expectedPostID, postID)
				}
			}
		})
	}
}

func TestTikTokPost_GetStatus(t *testing.T) {
	tests := []struct {
		name           string
		postID         string
		mockResponse   string
		mockStatusCode int
		expectedStatus string
		expectError    bool
	}{
		{
			name:           "published post status",
			postID:         "test_post_123",
			mockResponse:   `{"data":{"video_id":"vid123","post_status":"PUBLISHED","video_status":"PROCESSING_DONE"},"error":{"code":"","message":""}}`,
			mockStatusCode: 200,
			expectedStatus: "published",
			expectError:    false,
		},
		{
			name:           "processing post status",
			postID:         "test_post_123",
			mockResponse:   `{"data":{"video_id":"vid123","post_status":"PROCESSING","video_status":"PROCESSING"},"error":{"code":"","message":""}}`,
			mockStatusCode: 200,
			expectedStatus: "pending",
			expectError:    false,
		},
		{
			name:           "failed post status",
			postID:         "test_post_123",
			mockResponse:   `{"data":{"video_id":"vid123","post_status":"FAILED","fail_reason":"Content moderation failed"},"error":{"code":"","message":""}}`,
			mockStatusCode: 200,
			expectedStatus: "failed",
			expectError:    true,
		},
		{
			name:           "video status fallback",
			postID:         "test_post_123",
			mockResponse:   `{"data":{"video_id":"vid123","video_status":"UPLOADED"},"error":{"code":"","message":""}}`,
			mockStatusCode: 200,
			expectedStatus: "published",
			expectError:    false,
		},
		{
			name:           "API error response",
			postID:         "test_post_123",
			mockResponse:   `{"data":{},"error":{"code":"NOT_FOUND","message":"Post not found"}}`,
			mockStatusCode: 200,
			expectedStatus: "",
			expectError:    true,
		},
		{
			name:           "HTTP error",
			postID:         "test_post_123",
			mockResponse:   `{"error":"internal server error"}`,
			mockStatusCode: 500,
			expectedStatus: "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock HTTP client
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// Verify request
					if req.Method != "GET" {
						t.Errorf("Expected GET method, got %s", req.Method)
					}
					if req.Header.Get("Authorization") != "Bearer test_token" {
						t.Errorf("Expected Bearer token authorization")
					}

					// Return mock response
					return &http.Response{
						StatusCode: tt.mockStatusCode,
						Body:       io.NopCloser(bytes.NewBufferString(tt.mockResponse)),
					}, nil
				},
			}

			// Create Post
			config := &provider.ProviderConfig{
				AccessToken: "test_token",
				UserID:      "test_user",
			}
			Post := NewTikTokPost(config, mockClient)

			// Test GetStatus
			status, err := Post.GetStatus(context.Background(), tt.postID)

			// Verify results
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if status != tt.expectedStatus {
					t.Errorf("Expected status %s, got %s", tt.expectedStatus, status)
				}
			}
		})
	}
}

func TestTikTokPost_RefreshToken(t *testing.T) {
	tests := []struct {
		name           string
		refreshToken   string
		mockResponse   string
		mockStatusCode int
		expectError    bool
	}{
		{
			name:           "successful token refresh",
			refreshToken:   "refresh_token_123",
			mockResponse:   `{"data":{"access_token":"new_token","refresh_token":"new_refresh","expires_in":3600,"token_type":"Bearer","scope":"user.info.basic,video.upload,video.publish"},"error":{"code":"","message":""}}`,
			mockStatusCode: 200,
			expectError:    false,
		},
		{
			name:           "no refresh token",
			refreshToken:   "",
			mockResponse:   "",
			mockStatusCode: 200,
			expectError:    true,
		},
		{
			name:           "API error response",
			refreshToken:   "refresh_token_123",
			mockResponse:   `{"data":{},"error":{"code":"INVALID_TOKEN","message":"Invalid refresh token"}}`,
			mockStatusCode: 200,
			expectError:    true,
		},
		{
			name:           "HTTP error",
			refreshToken:   "refresh_token_123",
			mockResponse:   `{"error":"internal server error"}`,
			mockStatusCode: 500,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock HTTP client
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// Verify request
					if req.Method != "POST" {
						t.Errorf("Expected POST method, got %s", req.Method)
					}
					if req.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
						t.Errorf("Expected application/x-www-form-urlencoded content type")
					}

					// Return mock response
					return &http.Response{
						StatusCode: tt.mockStatusCode,
						Body:       io.NopCloser(bytes.NewBufferString(tt.mockResponse)),
					}, nil
				},
			}

			// Create Post
			config := &provider.ProviderConfig{
				AccessToken:  "old_token",
				RefreshToken: tt.refreshToken,
				UserID:       "test_user",
			}
			Post := NewTikTokPost(config, mockClient)

			// Test RefreshToken
			err := Post.RefreshToken(context.Background())

			// Verify results
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				// For successful refresh, check if token was updated
				if !tt.expectError && config.AccessToken != "new_token" {
					t.Errorf("Expected access token to be updated to 'new_token', got %s", config.AccessToken)
				}
			}
		})
	}
}
