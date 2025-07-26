package instagram

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
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

func TestInstagramPost_Publish(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		media          []provider.Media
		mockResponse   string
		mockStatusCode int
		expectedPostID string
		expectError    bool
	}{
		{
			name:           "successful publish with media",
			content:        "Test content",
			media:          []provider.Media{{FileName: "test.jpg", FilePath: "/tmp/test.jpg", MimeType: "image/jpeg"}},
			mockResponse:   `{"id":"instagram_container_123"}`,
			mockStatusCode: 200,
			expectedPostID: "instagram_post_456",
			expectError:    false,
		},
		{
			name:           "text-only post should fail",
			content:        "Test content",
			media:          []provider.Media{},
			mockResponse:   "",
			mockStatusCode: 200,
			expectedPostID: "",
			expectError:    true,
		},
		{
			name:           "API error response",
			content:        "Test content",
			media:          []provider.Media{{FileName: "test.jpg", FilePath: "/tmp/test.jpg", MimeType: "image/jpeg"}},
			mockResponse:   `{"error":{"message":"Invalid access token","type":"OAuthException","code":190}}`,
			mockStatusCode: 200,
			expectedPostID: "",
			expectError:    true,
		},
		{
			name:           "HTTP error",
			content:        "Test content",
			media:          []provider.Media{{FileName: "test.jpg", FilePath: "/tmp/test.jpg", MimeType: "image/jpeg"}},
			mockResponse:   `{"error":"internal server error"}`,
			mockStatusCode: 500,
			expectedPostID: "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip tests that require actual file operations for now
			if len(tt.media) > 0 && tt.media[0].FilePath != "" && !tt.expectError {
				t.Skip("Skipping test requiring file operations in mock environment")
				return
			}

			// Create mock HTTP client for Instagram Graph API
			callCount := 0
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					callCount++
					
					// For text-only posts, no HTTP calls should be made
					if len(tt.media) == 0 {
						t.Errorf("No HTTP calls expected for text-only posts")
					}

					// Verify Instagram Graph API URL structure
					if !strings.Contains(req.URL.String(), "graph.facebook.com/v18.0") {
						t.Errorf("Expected Instagram Graph API URL, got %s", req.URL.String())
					}

					// Verify request method
					if req.Method != "POST" {
						t.Errorf("Expected POST method, got %s", req.Method)
					}

					// Instagram Graph API uses multipart/form-data or form URL encoded, not JSON
					contentType := req.Header.Get("Content-Type")
					if !strings.Contains(contentType, "multipart/form-data") && !strings.Contains(contentType, "application/x-www-form-urlencoded") {
						t.Errorf("Expected multipart/form-data or form URL encoded, got %s", contentType)
					}

					// Instagram doesn't use Authorization header, token is in form data
					if req.Header.Get("Authorization") != "" {
						t.Errorf("Instagram API should not use Authorization header")
					}

					// Mock two-step process: first call creates container, second publishes
					if callCount == 1 {
						// First call: create media container
						return &http.Response{
							StatusCode: tt.mockStatusCode,
							Body:       io.NopCloser(bytes.NewBufferString(tt.mockResponse)),
						}, nil
					} else {
						// Second call: publish container
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewBufferString(`{"id":"instagram_post_456"}`)),
						}, nil
					}
				},
			}

			// Create Post
			config := &provider.ProviderConfig{
				AccessToken: "test_token",
				UserID:      "test_user",
			}
			Post := NewInstagramPost(config, mockClient)

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

func TestInstagramPost_GetStatus(t *testing.T) {
	tests := []struct {
		name           string
		postID         string
		mockResponse   string
		mockStatusCode int
		expectedStatus string
		expectError    bool
	}{
		{
			name:           "successful status check - published",
			postID:         "test_post_123",
			mockResponse:   `{"id":"test_post_123","media_type":"IMAGE","permalink":"https://instagram.com/p/test","timestamp":"2023-01-01T00:00:00Z"}`,
			mockStatusCode: 200,
			expectedStatus: "published",
			expectError:    false,
		},
		{
			name:           "successful status check - pending",
			postID:         "test_post_123",
			mockResponse:   `{"id":"test_post_123","media_type":"IMAGE","permalink":"","timestamp":""}`,
			mockStatusCode: 200,
			expectedStatus: "pending",
			expectError:    false,
		},
		{
			name:           "API error response",
			postID:         "test_post_123",
			mockResponse:   `{"error":{"message":"Post not found","type":"GraphMethodException","code":100}}`,
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

					// Verify Instagram Graph API URL structure
					if !strings.Contains(req.URL.String(), "graph.facebook.com/v18.0") {
						t.Errorf("Expected Instagram Graph API URL, got %s", req.URL.String())
					}

					// Verify access token is in URL parameters, not Authorization header
					if req.Header.Get("Authorization") != "" {
						t.Errorf("Instagram API should not use Authorization header")
					}
					if !strings.Contains(req.URL.String(), "access_token=test_token") {
						t.Errorf("Expected access_token in URL parameters")
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
			Post := NewInstagramPost(config, mockClient)

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

func TestInstagramPost_RefreshToken(t *testing.T) {
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
			mockResponse:   `{"access_token":"new_token","token_type":"bearer","expires_in":3600}`,
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
			mockResponse:   `{"error":{"message":"Invalid refresh token","type":"OAuthException","code":190}}`,
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
					// Verify request method - Instagram token refresh uses POST, not GET
					if req.Method != "POST" {
						t.Errorf("Expected POST method, got %s", req.Method)
					}

					// Verify Instagram Graph API URL structure
					if !strings.Contains(req.URL.String(), "graph.facebook.com/v18.0/oauth/access_token") {
						t.Errorf("Expected Instagram Graph API token refresh URL, got %s", req.URL.String())
					}

					// Verify content type for form data
					contentType := req.Header.Get("Content-Type")
					if !strings.Contains(contentType, "application/x-www-form-urlencoded") {
						t.Errorf("Expected form URL encoded content type, got %s", contentType)
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
			Post := NewInstagramPost(config, mockClient)

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
