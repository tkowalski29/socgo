package instagram

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	data_database "github.com/tkowalski/socgo/internal/data/database"
	"github.com/tkowalski/socgo/internal/data/provider"
)

// MockHTTPClient for testing
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestInstagramPost_validateInstagramConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *provider.ProviderConfig
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "empty UserID",
			config: &provider.ProviderConfig{
				UserID:      "",
				AccessToken: "test_token",
			},
			wantErr: true,
		},
		{
			name: "empty AccessToken",
			config: &provider.ProviderConfig{
				UserID:      "123456789",
				AccessToken: "",
			},
			wantErr: true,
		},
		{
			name: "invalid UserID format",
			config: &provider.ProviderConfig{
				UserID:      "abc123",
				AccessToken: "test_token",
			},
			wantErr: true,
		},
		{
			name: "valid config",
			config: &provider.ProviderConfig{
				UserID:      "123456789012345",
				AccessToken: "test_token",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &InstagramPost{
				Config: tt.config,
			}
			
			err := p.validateInstagramConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("validateInstagramConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidInstagramBusinessAccountID(t *testing.T) {
	tests := []struct {
		name   string
		userID string
		want   bool
	}{
		{
			name:   "too short",
			userID: "123",
			want:   false,
		},
		{
			name:   "too long",
			userID: "12345678901234567890123456",
			want:   false,
		},
		{
			name:   "contains letters",
			userID: "123abc456",
			want:   false,
		},
		{
			name:   "valid short ID",
			userID: "12345678",
			want:   true,
		},
		{
			name:   "valid long ID",
			userID: "1234567890123456789012345",
			want:   true,
		},
		{
			name:   "typical Instagram Business ID",
			userID: "123456789012345",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidInstagramBusinessAccountID(tt.userID)
			if got != tt.want {
				t.Errorf("isValidInstagramBusinessAccountID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInstagramPost_checkInstagramPermissions(t *testing.T) {
	tests := []struct {
		name         string
		userID       string
		response     string
		statusCode   int
		wantErr      bool
		expectedType string
	}{
		{
			name:       "business account success",
			userID:     "123456789",
			statusCode: 200,
			response: `{
				"id": "123456789",
				"username": "testuser",
				"account_type": "BUSINESS",
				"media_count": 10
			}`,
			wantErr:      false,
			expectedType: "BUSINESS",
		},
		{
			name:       "creator account success",
			userID:     "123456789",
			statusCode: 200,
			response: `{
				"id": "123456789",
				"username": "testuser",
				"account_type": "CREATOR",
				"media_count": 5
			}`,
			wantErr:      false,
			expectedType: "CREATOR",
		},
		{
			name:       "personal account error",
			userID:     "123456789",
			statusCode: 200,
			response: `{
				"id": "123456789",
				"username": "testuser",
				"account_type": "PERSONAL",
				"media_count": 3
			}`,
			wantErr:      true,
			expectedType: "PERSONAL",
		},
		{
			name:       "API error",
			userID:     "123456789",
			statusCode: 400,
			response: `{
				"error": {
					"message": "Invalid access token",
					"type": "OAuthException",
					"code": 190
				}
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// Verify the request
					expectedURL := "https://graph.facebook.com/v18.0/" + tt.userID
					if !strings.Contains(req.URL.String(), expectedURL) {
						t.Errorf("Expected URL to contain %s, got %s", expectedURL, req.URL.String())
					}

					// Return mock response
					resp := &http.Response{
						StatusCode: tt.statusCode,
						Body:       io.NopCloser(strings.NewReader(tt.response)),
						Header:     make(http.Header),
					}
					resp.Header.Set("Content-Type", "application/json")
					return resp, nil
				},
			}

			p := &InstagramPost{
				Config: &provider.ProviderConfig{
					UserID:      tt.userID,
					AccessToken: "test_token",
				},
				HttpClient: mockClient,
			}

			err := p.checkInstagramPermissions(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("checkInstagramPermissions() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.expectedType != "" && !strings.Contains(tt.response, tt.expectedType) {
				t.Errorf("Expected account type %s in response", tt.expectedType)
			}
		})
	}
}

func TestInstagramPost_Publish_NoMedia(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Mock response for Instagram permissions check
			resp := &http.Response{
				StatusCode: 200,
				Body: io.NopCloser(strings.NewReader(`{
					"id": "123456789012345",
					"username": "testuser",
					"account_type": "BUSINESS",
					"media_count": 10
				}`)),
				Header: make(http.Header),
			}
			resp.Header.Set("Content-Type", "application/json")
			return resp, nil
		},
	}

	p := &InstagramPost{
		Config: &provider.ProviderConfig{
			UserID:      "123456789012345",
			AccessToken: "test_token",
		},
		HttpClient: mockClient,
	}

	// Create a mock database provider
	mockDB := data_database.Provider{
		ID:     1,
		Name:   "Test Provider",
		Type:   "instagram",
		UserID: "test_user",
	}

	_, err := p.Publish(context.Background(), mockDB, "Test content", nil)
	if err == nil {
		t.Error("Expected error for post without media")
	}

	if !strings.Contains(err.Error(), "require at least one media item") {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestInstagramPost_createMediaContainer(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Verify request method and URL
			if req.Method != "POST" {
				t.Errorf("Expected POST request, got %s", req.Method)
			}

			expectedURL := "https://graph.facebook.com/v18.0/123456789/media"
			if req.URL.String() != expectedURL {
				t.Errorf("Expected URL %s, got %s", expectedURL, req.URL.String())
			}

			// Mock successful response
			resp := &http.Response{
				StatusCode: 200,
				Body: io.NopCloser(strings.NewReader(`{
					"id": "container_123456"
				}`)),
				Header: make(http.Header),
			}
			resp.Header.Set("Content-Type", "application/json")
			return resp, nil
		},
	}

	p := &InstagramPost{
		Config: &provider.ProviderConfig{
			UserID:      "123456789",
			AccessToken: "test_token",
		},
		HttpClient: mockClient,
	}

	// Create a temporary test file path
	testFile := "/tmp/test_image.jpg"
	
	// For this test, we'll mock the file operations
	// In a real test, you'd create an actual test file
	
	media := provider.Media{
		FilePath: testFile,
		FileName: "test_image.jpg",
		MimeType: "image/jpeg",
	}

	// This test will fail because the file doesn't exist
	// In a proper test setup, you'd create test files
	_, err := p.createMediaContainer(context.Background(), "Test caption", media, "")
	
	// We expect an error because the file doesn't exist
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestInstagramPost_GetStatus(t *testing.T) {
	tests := []struct {
		name       string
		postID     string
		response   string
		statusCode int
		wantStatus string
		wantErr    bool
	}{
		{
			name:       "published post",
			postID:     "123456789",
			statusCode: 200,
			response: `{
				"id": "123456789",
				"media_type": "IMAGE",
				"permalink": "https://instagram.com/p/abc123",
				"timestamp": "2024-01-01T00:00:00Z"
			}`,
			wantStatus: string(provider.PostStatusPublished),
			wantErr:    false,
		},
		{
			name:       "API error",
			postID:     "123456789",
			statusCode: 400,
			response: `{
				"error": {
					"message": "Post not found",
					"type": "OAuthException",
					"code": 100
				}
			}`,
			wantStatus: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					resp := &http.Response{
						StatusCode: tt.statusCode,
						Body:       io.NopCloser(strings.NewReader(tt.response)),
						Header:     make(http.Header),
					}
					resp.Header.Set("Content-Type", "application/json")
					return resp, nil
				},
			}

			p := &InstagramPost{
				Config: &provider.ProviderConfig{
					UserID:      "123456789",
					AccessToken: "test_token",
				},
				HttpClient: mockClient,
			}

			status, err := p.GetStatus(context.Background(), tt.postID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStatus() error = %v, wantErr %v", err, tt.wantErr)
			}

			if status != tt.wantStatus {
				t.Errorf("GetStatus() status = %v, want %v", status, tt.wantStatus)
			}
		})
	}
}