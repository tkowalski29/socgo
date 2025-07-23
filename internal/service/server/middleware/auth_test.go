package middleware

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tkowalski/socgo/internal/database"
)

func TestAuthMiddleware_APIAuthMiddleware_ValidToken(t *testing.T) {
	// Create test database manager
	dbManager := database.NewTestManager(t)
	defer dbManager.Close()

	// Create a test token
	token := "test-token-12345"
	tokenHash := sha256.Sum256([]byte(token))
	tokenHashString := fmt.Sprintf("%x", tokenHash)

	// Save token to database
	db, err := dbManager.GetDB("default_user")
	if err != nil {
		t.Fatalf("Failed to get database: %v", err)
	}

	apiToken := database.APIToken{
		Hash:      tokenHashString,
		UserID:    "default_user",
		CreatedAt: time.Now(),
	}

	if err := db.Create(&apiToken).Error; err != nil {
		t.Fatalf("Failed to create API token: %v", err)
	}

	// Create middleware
	authMiddleware := NewAuthMiddleware(dbManager)

	// Create test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("success")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	// Wrap with middleware
	handler := authMiddleware.APIAuthMiddleware(testHandler)

	// Create request with valid token
	req, err := http.NewRequest("GET", "/api/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if rr.Body.String() != "success" {
		t.Errorf("Handler returned unexpected body: got %v want %v", rr.Body.String(), "success")
	}

	// Verify last_used was updated
	var updatedToken database.APIToken
	if err := db.First(&updatedToken, apiToken.ID).Error; err != nil {
		t.Fatalf("Failed to retrieve updated token: %v", err)
	}

	if updatedToken.LastUsed == nil {
		t.Error("Expected last_used to be updated")
	}
}

func TestAuthMiddleware_APIAuthMiddleware_NoAuthHeader(t *testing.T) {
	dbManager := database.NewTestManager(t)
	defer dbManager.Close()

	authMiddleware := NewAuthMiddleware(dbManager)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("success")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	handler := authMiddleware.APIAuthMiddleware(testHandler)

	req, err := http.NewRequest("GET", "/api/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_APIAuthMiddleware_InvalidToken(t *testing.T) {
	dbManager := database.NewTestManager(t)
	defer dbManager.Close()

	authMiddleware := NewAuthMiddleware(dbManager)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("success")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	handler := authMiddleware.APIAuthMiddleware(testHandler)

	req, err := http.NewRequest("GET", "/api/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer invalid-token")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_APIAuthMiddleware_InvalidAuthFormat(t *testing.T) {
	dbManager := database.NewTestManager(t)
	defer dbManager.Close()

	authMiddleware := NewAuthMiddleware(dbManager)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("success")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	})

	handler := authMiddleware.APIAuthMiddleware(testHandler)

	req, err := http.NewRequest("GET", "/api/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Basic sometoken")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}
