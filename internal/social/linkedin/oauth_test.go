package linkedin

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/tkowalski/socgo/internal/data/config"
	oauth_data "github.com/tkowalski/socgo/internal/data/oauth"
)

// MockHTTPClient for testing
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestLinkedInOAuth_GetProviderType(t *testing.T) {
	cfg := &config.Config{}
	oauth := NewOAuth(cfg)

	providerType := oauth.GetProviderType()
	expected := oauth_data.ProviderTypeLinkedIn

	if providerType != expected {
		t.Errorf("Expected provider type %s, got %s", expected, providerType)
	}
}

func TestLinkedInOAuth_GetConnectURL(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			BaseURL: "https://example.com",
		},
	}

	oauth := NewOAuth(cfg)

	providerConfig := &config.ProviderInstance{
		ClientID:     "test_client_id",
		ClientSecret: "test_client_secret",
	}

	connectURL, err := oauth.GetConnectURL("test_user", "test_provider", providerConfig)
	if err != nil {
		t.Fatalf("GetConnectURL failed: %v", err)
	}

	if !strings.Contains(connectURL, "https://www.linkedin.com/oauth/v2/authorization") {
		t.Errorf("Expected LinkedIn OAuth URL, got: %s", connectURL)
	}

	if !strings.Contains(connectURL, "client_id=test_client_id") {
		t.Errorf("Expected client_id in URL, got: %s", connectURL)
	}

	if !strings.Contains(connectURL, "scope=openid+profile+w_member_social+email") {
		t.Errorf("Expected correct scopes in URL, got: %s", connectURL)
	}
}

func TestLinkedInOAuth_ExchangeCodeForToken(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			BaseURL: "https://example.com",
		},
	}

	providerConfig := &config.ProviderInstance{
		ClientID:     "test_client_id",
		ClientSecret: "test_client_secret",
	}

	// Mock token response
	tokenResponse := map[string]interface{}{
		"access_token":  "test_access_token",
		"token_type":    "Bearer",
		"expires_in":    3600,
		"refresh_token": "test_refresh_token",
		"scope":         "openid profile w_member_social email",
	}

	responseBody, _ := json.Marshal(tokenResponse)

	// Create mock HTTP client
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			resp := &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(string(responseBody))),
				Header:     make(http.Header),
			}
			resp.Header.Set("Content-Type", "application/json")
			return resp, nil
		},
	}

	oauth := NewOAuthWithClient(cfg, mockClient)

	config, err := oauth.ExchangeCodeForToken("test_code", providerConfig)
	if err != nil {
		t.Fatalf("ExchangeCodeForToken failed: %v", err)
	}

	if config.AccessToken != "test_access_token" {
		t.Errorf("Expected access token 'test_access_token', got: %s", config.AccessToken)
	}

	if config.TokenType != "Bearer" {
		t.Errorf("Expected token type 'Bearer', got: %s", config.TokenType)
	}
}

func TestLinkedInOAuth_GetUserInfo(t *testing.T) {
	cfg := &config.Config{}

	// Mock user info response
	userInfoResponse := map[string]interface{}{
		"sub":            "linkedin_user_123",
		"name":           "John Doe",
		"given_name":     "John",
		"family_name":    "Doe",
		"picture":        "https://example.com/avatar.jpg",
		"email":          "john.doe@example.com",
		"email_verified": true,
	}

	responseBody, _ := json.Marshal(userInfoResponse)

	// Create mock HTTP client
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			if req.URL.String() != "https://api.linkedin.com/v2/userinfo" {
				t.Errorf("Expected LinkedIn userinfo URL, got: %s", req.URL.String())
			}

			if req.Header.Get("Authorization") != "Bearer test_token" {
				t.Errorf("Expected Authorization header with bearer token")
			}

			resp := &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(string(responseBody))),
				Header:     make(http.Header),
			}
			resp.Header.Set("Content-Type", "application/json")
			return resp, nil
		},
	}

	oauth := NewOAuthWithClient(cfg, mockClient)
	userInfo, err := oauth.GetUserInfo("test_token")
	if err != nil {
		t.Fatalf("GetUserInfo failed: %v", err)
	}

	if userInfo.ID != "linkedin_user_123" {
		t.Errorf("Expected user ID 'linkedin_user_123', got: %s", userInfo.ID)
	}

	if userInfo.Name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got: %s", userInfo.Name)
	}

	if userInfo.Email != "john.doe@example.com" {
		t.Errorf("Expected email 'john.doe@example.com', got: %s", userInfo.Email)
	}
}

func TestLinkedInOAuth_GetAvailableAccounts(t *testing.T) {
	cfg := &config.Config{}

	// For LinkedIn, GetAvailableAccounts calls GetUserInfo
	// Mock user info response
	userInfoResponse := map[string]interface{}{
		"sub":   "linkedin_user_123",
		"name":  "John Doe",
		"email": "john.doe@example.com",
	}

	responseBody, _ := json.Marshal(userInfoResponse)

	// Create mock HTTP client
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			resp := &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(string(responseBody))),
				Header:     make(http.Header),
			}
			resp.Header.Set("Content-Type", "application/json")
			return resp, nil
		},
	}

	oauth := NewOAuthWithClient(cfg, mockClient)
	accounts, err := oauth.GetAvailableAccounts("test_token")
	if err != nil {
		t.Fatalf("GetAvailableAccounts failed: %v", err)
	}

	if len(accounts) != 1 {
		t.Errorf("Expected 1 account, got %d", len(accounts))
	}

	account := accounts[0]
	if account.ID != "linkedin_user_123" {
		t.Errorf("Expected account ID 'linkedin_user_123', got: %s", account.ID)
	}

	if account.Name != "John Doe" {
		t.Errorf("Expected account name 'John Doe', got: %s", account.Name)
	}

	if account.Type != "personal" {
		t.Errorf("Expected account type 'personal', got: %s", account.Type)
	}
}

func TestLinkedInOAuth_SaveAllAccounts(t *testing.T) {
	cfg := &config.Config{}

	token := &oauth_data.ProviderConfig{
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
		TokenType:    "Bearer",
	}

	// Mock GetAvailableAccounts to return one account
	userInfoResponse := map[string]interface{}{
		"sub":   "linkedin_user_123",
		"name":  "John Doe",
		"email": "john.doe@example.com",
	}

	responseBody, _ := json.Marshal(userInfoResponse)

	// Create mock HTTP client
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			resp := &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(string(responseBody))),
				Header:     make(http.Header),
			}
			resp.Header.Set("Content-Type", "application/json")
			return resp, nil
		},
	}

	oauth := NewOAuthWithClient(cfg, mockClient)

	// Mock saveFunc
	var savedProviders []string
	saveFunc := func(userID string, providerType oauth_data.ProviderType, providerName string, config *oauth_data.ProviderConfig) error {
		savedProviders = append(savedProviders, providerName)

		if providerType != oauth_data.ProviderTypeLinkedIn {
			t.Errorf("Expected provider type LinkedIn, got: %s", providerType)
		}

		if config.AccessToken != "test_access_token" {
			t.Errorf("Expected access token to be passed through")
		}

		return nil
	}

	err := oauth.SaveAllAccounts("test_user", "test_provider", token, saveFunc)
	if err != nil {
		t.Fatalf("SaveAllAccounts failed: %v", err)
	}

	if len(savedProviders) != 1 {
		t.Errorf("Expected 1 saved provider, got %d", len(savedProviders))
	}

	if savedProviders[0] != "test_provider_John Doe" {
		t.Errorf("Expected provider name 'test_provider_John Doe', got: %s", savedProviders[0])
	}
}

// mockTransport is a helper for mocking HTTP transport
type mockTransport struct {
	client *MockHTTPClient
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.client.Do(req)
}
