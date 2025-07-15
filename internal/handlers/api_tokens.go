package handlers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/tkowalski/socgo/internal/database"
)

type APITokenHandler struct {
	dbManager *database.Manager
	secretKey []byte
}

type APITokenRequest struct {
	// Token generation doesn't require additional parameters
}

type APITokenResponse struct {
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	Message   string    `json:"message"`
}

func NewAPITokenHandler(dbManager *database.Manager) *APITokenHandler {
	// In production, this should come from environment variables
	secretKey := []byte("your-secret-key-change-in-production")
	return &APITokenHandler{
		dbManager: dbManager,
		secretKey: secretKey,
	}
}

func (h *APITokenHandler) HandleCreateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID (currently defaults to "default_user")
	userID := h.getUserID(r)

	// Generate random bytes for token
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		log.Printf("Error generating random bytes: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create token payload with user ID and timestamp
	payload := fmt.Sprintf("%s:%d:%s", userID, time.Now().Unix(), base64.URLEncoding.EncodeToString(randomBytes))
	
	// Generate HMAC-SHA256 signature
	mac := hmac.New(sha256.New, h.secretKey)
	mac.Write([]byte(payload))
	signature := mac.Sum(nil)
	
	// Combine payload and signature for the final token
	token := base64.URLEncoding.EncodeToString([]byte(payload + ":" + base64.URLEncoding.EncodeToString(signature)))
	
	// Generate hash for storage (SHA256 of the token)
	tokenHash := sha256.Sum256([]byte(token))
	tokenHashString := fmt.Sprintf("%x", tokenHash)

	// Get database instance for user
	db, err := h.dbManager.GetDB(userID)
	if err != nil {
		log.Printf("Error getting database: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Save token hash to database
	apiToken := database.APIToken{
		Hash:      tokenHashString,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	if err := db.Create(&apiToken).Error; err != nil {
		log.Printf("Error saving API token: %v", err)
		http.Error(w, "Failed to create API token", http.StatusInternalServerError)
		return
	}

	// Return the token (only once)
	response := APITokenResponse{
		Token:     token,
		CreatedAt: apiToken.CreatedAt,
		Message:   "API token created successfully. Store it securely as it won't be shown again.",
	}

	h.writeJSONResponse(w, response, http.StatusCreated)
}

func (h *APITokenHandler) getUserID(r *http.Request) string {
	// TODO: Implement proper user authentication
	return "default_user"
}

func (h *APITokenHandler) writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}