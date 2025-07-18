package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/tkowalski/socgo/internal/config"
	"github.com/tkowalski/socgo/internal/database"
	"github.com/tkowalski/socgo/internal/middleware"
	"github.com/tkowalski/socgo/internal/oauth"
	"github.com/tkowalski/socgo/internal/providers"
)

func TestAPITokenIntegration_EndToEnd(t *testing.T) {
	// Create test database manager
	dbManager := database.NewTestManager(t)
	defer dbManager.Close()

	// Create handlers and middleware
	apiTokenHandler := NewAPITokenHandler(dbManager)
	authMiddleware := middleware.NewAuthMiddleware(dbManager)

	// Create oauth service and provider service for testing
	cfg := &config.Config{}
	oauthService := oauth.NewService(dbManager, cfg)
	providerService := providers.NewProviderService(dbManager, oauthService)
	postHandler := NewPostHandler(dbManager, providerService)

	// Create router with routes similar to server setup
	r := mux.NewRouter()

	// API token generation endpoint (public)
	r.HandleFunc("/api-tokens", apiTokenHandler.HandleCreateToken).Methods("POST")

	// Protected API routes with auth middleware
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.Use(authMiddleware.APIAuthMiddleware)
	apiRouter.HandleFunc("/posts", postHandler.HandlePost).Methods("POST")

	// Test 1: Generate API token
	tokenReq, err := http.NewRequest("POST", "/api-tokens", bytes.NewBuffer([]byte("{}")))
	if err != nil {
		t.Fatal(err)
	}
	tokenReq.Header.Set("Content-Type", "application/json")

	tokenRR := httptest.NewRecorder()
	r.ServeHTTP(tokenRR, tokenReq)

	if status := tokenRR.Code; status != http.StatusCreated {
		t.Errorf("Token creation failed: got %v want %v", status, http.StatusCreated)
	}

	var tokenResponse APITokenResponse
	if err := json.Unmarshal(tokenRR.Body.Bytes(), &tokenResponse); err != nil {
		t.Fatalf("Failed to unmarshal token response: %v", err)
	}

	if tokenResponse.Token == "" {
		t.Fatal("Expected token to be returned")
	}

	// Test 2: Use token to access protected endpoint (should fail without proper provider setup)
	postReq, err := http.NewRequest("POST", "/api/posts", bytes.NewBuffer([]byte(`{"provider_id": 1, "content": "test post"}`)))
	if err != nil {
		t.Fatal(err)
	}
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("Authorization", "Bearer "+tokenResponse.Token)

	postRR := httptest.NewRecorder()
	r.ServeHTTP(postRR, postReq)

	// Should get 404 for provider (which is expected since no provider is configured)
	if status := postRR.Code; status != http.StatusNotFound {
		t.Errorf("Expected 404 for missing provider, got %v", status)
	}

	// Test 3: Try to access protected endpoint without token
	postReqNoAuth, err := http.NewRequest("POST", "/api/posts", bytes.NewBuffer([]byte(`{"provider_id": 1, "content": "test post"}`)))
	if err != nil {
		t.Fatal(err)
	}
	postReqNoAuth.Header.Set("Content-Type", "application/json")

	postRRNoAuth := httptest.NewRecorder()
	r.ServeHTTP(postRRNoAuth, postReqNoAuth)

	if status := postRRNoAuth.Code; status != http.StatusUnauthorized {
		t.Errorf("Expected 401 without token, got %v", status)
	}

	// Test 4: Try to access protected endpoint with invalid token
	postReqInvalidAuth, err := http.NewRequest("POST", "/api/posts", bytes.NewBuffer([]byte(`{"provider_id": 1, "content": "test post"}`)))
	if err != nil {
		t.Fatal(err)
	}
	postReqInvalidAuth.Header.Set("Content-Type", "application/json")
	postReqInvalidAuth.Header.Set("Authorization", "Bearer invalid-token")

	postRRInvalidAuth := httptest.NewRecorder()
	r.ServeHTTP(postRRInvalidAuth, postReqInvalidAuth)

	if status := postRRInvalidAuth.Code; status != http.StatusUnauthorized {
		t.Errorf("Expected 401 with invalid token, got %v", status)
	}
}
