package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tkowalski/socgo/internal/database"
)

func TestAPITokenHandler_HandleCreateToken(t *testing.T) {
	// Create test database manager
	dbManager := database.NewTestManager(t)
	defer dbManager.Close()

	// Create handler
	handler := NewAPITokenHandler(dbManager)

	// Test POST request
	req, err := http.NewRequest("POST", "/api-tokens", bytes.NewBuffer([]byte("{}")))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.HandleCreateToken(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	// Check response
	var response APITokenResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Token == "" {
		t.Error("Expected token to be returned")
	}

	if response.CreatedAt.IsZero() {
		t.Error("Expected created_at to be set")
	}

	if response.Message == "" {
		t.Error("Expected message to be returned")
	}

	// Verify token was saved to database
	db, err := dbManager.GetDB("default_user")
	if err != nil {
		t.Fatalf("Failed to get database: %v", err)
	}

	var tokenCount int64
	if err := db.Model(&database.APIToken{}).Count(&tokenCount).Error; err != nil {
		t.Fatalf("Failed to count tokens: %v", err)
	}

	if tokenCount != 1 {
		t.Errorf("Expected 1 token in database, got %d", tokenCount)
	}
}

func TestAPITokenHandler_HandleCreateToken_InvalidMethod(t *testing.T) {
	dbManager := database.NewTestManager(t)
	defer dbManager.Close()

	handler := NewAPITokenHandler(dbManager)

	// Test GET request (should fail)
	req, err := http.NewRequest("GET", "/api-tokens", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.HandleCreateToken(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}