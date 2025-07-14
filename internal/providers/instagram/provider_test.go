package instagram

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/tkowalski/socgo/internal/providers"
)

// MockHTTPClient is a mock implementation of HTTPClient for testing
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestInstagramProvider_Publish(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		mockResponse   string
		mockStatusCode int
		expectedPostID string
		expectError    bool
	}{
		{
			name:           "successful publish",
			content:        "Test content",
			mockResponse:   `{"id":"instagram_post_123"}`,
			mockStatusCode: 200,
			expectedPostID: "instagram_post_123",
			expectError:    false,
		},
		{
			name:           "successful publish with fallback postID",
			content:        "Test content",
			mockResponse:   `{"id":""}`,
			mockStatusCode: 200,
			expectedPostID: "instagram_", // prefix only, timestamp will vary
			expectError:    false,
		},
		{
			name:           "API error response",
			content:        "Test content",
			mockResponse:   `{"error":{"message":"Invalid access token","type":"OAuthException","code":190}}`,
			mockStatusCode: 200,
			expectedPostID: "",
			expectError:    true,
		},
		{
			name:           "HTTP error",
			content:        "Test content",
			mockResponse:   `{"error":"internal server error"}`,
			mockStatusCode: 500,
			expectedPostID: "",
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
					if req.Header.Get("Content-Type") != "application/json" {
						t.Errorf("Expected application/json content type")
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

			// Create provider
			config := &providers.ProviderConfig{
				AccessToken: "test_token",
				UserID:      "test_user",
			}
			provider := NewInstagramProvider(config, mockClient)

			// Test Publish
			postID, err := provider.Publish(context.Background(), tt.content)

			// Verify results
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if tt.expectedPostID != "instagram_" {
					if postID != tt.expectedPostID {
						t.Errorf("Expected postID %s, got %s", tt.expectedPostID, postID)
					}
				} else {
					// Check prefix for generated postID
					if len(postID) < 10 || postID[:10] != "instagram_" {
						t.Errorf("Expected postID to start with 'instagram_', got %s", postID)
					}
				}
			}
		})
	}
}

func TestInstagramProvider_GetStatus(t *testing.T) {
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
			mockResponse:   `{"id":"test_post_123","media_type":"IMAGE","permalink":"","timestamp":"2023-01-01T00:00:00Z"}`,
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

			// Create provider
			config := &providers.ProviderConfig{
				AccessToken: "test_token",
				UserID:      "test_user",
			}
			provider := NewInstagramProvider(config, mockClient)

			// Test GetStatus
			status, err := provider.GetStatus(context.Background(), tt.postID)

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

func TestInstagramProvider_RefreshToken(t *testing.T) {
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
					// Verify request
					if req.Method != "GET" {
						t.Errorf("Expected GET method, got %s", req.Method)
					}
					if req.Header.Get("Content-Type") != "application/json" {
						t.Errorf("Expected application/json content type")
					}

					// Return mock response
					return &http.Response{
						StatusCode: tt.mockStatusCode,
						Body:       io.NopCloser(bytes.NewBufferString(tt.mockResponse)),
					}, nil
				},
			}

			// Create provider
			config := &providers.ProviderConfig{
				AccessToken:  "old_token",
				RefreshToken: tt.refreshToken,
				UserID:       "test_user",
			}
			provider := NewInstagramProvider(config, mockClient)

			// Test RefreshToken
			err := provider.RefreshToken(context.Background())

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
