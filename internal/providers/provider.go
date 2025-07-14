package providers

import (
	"context"
	"net/http"
)

// Provider defines the interface for social media providers
type Provider interface {
	// Publish publishes content to the social media platform
	// Returns postID on success
	Publish(ctx context.Context, content string) (postID string, err error)
	
	// GetStatus retrieves the status of a published post
	GetStatus(ctx context.Context, postID string) (status string, err error)
	
	// RefreshToken refreshes the access token if needed
	RefreshToken(ctx context.Context) error
}

// HTTPClient interface for mocking HTTP requests in tests
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// ProviderConfig contains configuration for a provider
type ProviderConfig struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresAt    int64  `json:"expires_at"`
	Scope        string `json:"scope,omitempty"`
	UserID       string `json:"user_id,omitempty"`
}

// PostStatus represents possible post statuses
type PostStatus string

const (
	PostStatusPublished PostStatus = "published"
	PostStatusPending   PostStatus = "pending"
	PostStatusFailed    PostStatus = "failed"
	PostStatusDeleted   PostStatus = "deleted"
)