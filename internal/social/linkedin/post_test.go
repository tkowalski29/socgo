package linkedin

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	data_database "github.com/tkowalski/socgo/internal/data/database"
	"github.com/tkowalski/socgo/internal/data/provider"
)

func TestLinkedInPost_validateLinkedInConfig(t *testing.T) {
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
				AccessToken: "test_token",
			},
			wantErr: true,
		},
		{
			name: "empty AccessToken",
			config: &provider.ProviderConfig{
				UserID: "test_user",
			},
			wantErr: true,
		},
		{
			name: "valid config",
			config: &provider.ProviderConfig{
				UserID:      "test_user",
				AccessToken: "test_token",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &LinkedInPost{
				Config: tt.config,
			}
			
			err := p.validateLinkedInConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("validateLinkedInConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLinkedInPost_publishTextOnly(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Verify request method and URL
			if req.Method != "POST" {
				t.Errorf("Expected POST request, got %s", req.Method)
			}

			expectedURL := "https://api.linkedin.com/v2/ugcPosts"
			if req.URL.String() != expectedURL {
				t.Errorf("Expected URL %s, got %s", expectedURL, req.URL.String())
			}

			// Verify headers
			if req.Header.Get("Authorization") != "Bearer test_token" {
				t.Errorf("Expected Authorization header with bearer token")
			}

			if req.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected Content-Type application/json")
			}

			if req.Header.Get("X-Restli-Protocol-Version") != "2.0.0" {
				t.Errorf("Expected X-Restli-Protocol-Version 2.0.0")
			}

			// Mock successful response
			resp := &http.Response{
				StatusCode: 201, // LinkedIn returns 201 for created posts
				Body: io.NopCloser(strings.NewReader(`{
					"id": "urn:li:ugcPost:123456789"
				}`)),
				Header: make(http.Header),
			}
			resp.Header.Set("Content-Type", "application/json")
			return resp, nil
		},
	}

	p := &LinkedInPost{
		Config: &provider.ProviderConfig{
			UserID:      "test_user_123",
			AccessToken: "test_token",
		},
		HttpClient: mockClient,
		BaseURL:    "https://test.example.com",
	}

	postID, err := p.publishTextOnly(context.Background(), "Test content")
	if err != nil {
		t.Fatalf("publishTextOnly failed: %v", err)
	}

	if postID != "urn:li:ugcPost:123456789" {
		t.Errorf("Expected post ID 'urn:li:ugcPost:123456789', got: %s", postID)
	}
}

func TestLinkedInPost_Publish_NoMedia(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Mock successful response for text-only post
			resp := &http.Response{
				StatusCode: 201,
				Body: io.NopCloser(strings.NewReader(`{
					"id": "urn:li:ugcPost:123456789"
				}`)),
				Header: make(http.Header),
			}
			resp.Header.Set("Content-Type", "application/json")
			return resp, nil
		},
	}

	p := &LinkedInPost{
		Config: &provider.ProviderConfig{
			UserID:      "test_user_123",
			AccessToken: "test_token",
		},
		HttpClient: mockClient,
		BaseURL:    "https://test.example.com",
	}

	// Create a mock database provider
	mockDB := data_database.Provider{
		ID:     1,
		Name:   "Test Provider",
		Type:   "linkedin",
		UserID: "test_user",
	}

	postID, err := p.Publish(context.Background(), mockDB, "Test content", nil)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	if postID != "urn:li:ugcPost:123456789" {
		t.Errorf("Expected post ID 'urn:li:ugcPost:123456789', got: %s", postID)
	}
}

func TestLinkedInPost_Publish_WithMedia(t *testing.T) {
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Mock successful response for post with media (currently falls back to text-only)
			resp := &http.Response{
				StatusCode: 201,
				Body: io.NopCloser(strings.NewReader(`{
					"id": "urn:li:ugcPost:123456789"
				}`)),
				Header: make(http.Header),
			}
			resp.Header.Set("Content-Type", "application/json")
			return resp, nil
		},
	}

	p := &LinkedInPost{
		Config: &provider.ProviderConfig{
			UserID:      "test_user_123",
			AccessToken: "test_token",
		},
		HttpClient: mockClient,
		BaseURL:    "https://test.example.com",
	}

	// Create a mock database provider
	mockDB := data_database.Provider{
		ID:     1,
		Name:   "Test Provider",
		Type:   "linkedin",
		UserID: "test_user",
	}

	media := []provider.Media{
		{
			FileName: "test_image.jpg",
			FilePath: "/tmp/test_image.jpg",
			MimeType: "image/jpeg",
		},
	}

	postID, err := p.Publish(context.Background(), mockDB, "Test content with media", media)
	if err != nil {
		t.Fatalf("Publish with media failed: %v", err)
	}

	if postID != "urn:li:ugcPost:123456789" {
		t.Errorf("Expected post ID 'urn:li:ugcPost:123456789', got: %s", postID)
	}
}

func TestLinkedInPost_GetStatus(t *testing.T) {
	tests := []struct {
		name         string
		postID       string
		response     string
		statusCode   int
		expectedStatus string
		wantErr      bool
	}{
		{
			name:   "published post",
			postID: "urn:li:ugcPost:123456789",
			response: `{
				"id": "urn:li:ugcPost:123456789",
				"lifecycleState": "PUBLISHED"
			}`,
			statusCode:     200,
			expectedStatus: "published",
			wantErr:        false,
		},
		{
			name:   "draft post",
			postID: "urn:li:ugcPost:123456789",
			response: `{
				"id": "urn:li:ugcPost:123456789",
				"lifecycleState": "DRAFT"
			}`,
			statusCode:     200,
			expectedStatus: "pending",
			wantErr:        false,
		},
		{
			name:   "API error",
			postID: "urn:li:ugcPost:123456789",
			response: `{
				"error": {
					"message": "Invalid post ID",
					"status": 404
				}
			}`,
			statusCode: 200,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					expectedURL := "https://api.linkedin.com/v2/ugcPosts/" + tt.postID
					if req.URL.String() != expectedURL {
						t.Errorf("Expected URL %s, got %s", expectedURL, req.URL.String())
					}

					resp := &http.Response{
						StatusCode: tt.statusCode,
						Body:       io.NopCloser(strings.NewReader(tt.response)),
						Header:     make(http.Header),
					}
					resp.Header.Set("Content-Type", "application/json")
					return resp, nil
				},
			}

			p := &LinkedInPost{
				Config: &provider.ProviderConfig{
					UserID:      "test_user_123",
					AccessToken: "test_token",
				},
				HttpClient: mockClient,
				BaseURL:    "https://test.example.com",
			}

			status, err := p.GetStatus(context.Background(), tt.postID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && status != tt.expectedStatus {
				t.Errorf("GetStatus() status = %v, want %v", status, tt.expectedStatus)
			}
		})
	}
}