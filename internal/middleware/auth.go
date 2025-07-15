package middleware

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/tkowalski/socgo/internal/database"
)

type AuthMiddleware struct {
	dbManager *database.Manager
}

func NewAuthMiddleware(dbManager *database.Manager) *AuthMiddleware {
	return &AuthMiddleware{
		dbManager: dbManager,
	}
}

// APIAuthMiddleware checks for valid Bearer token in Authorization header
func (m *AuthMiddleware) APIAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.writeUnauthorizedResponse(w, "Authorization header required")
			return
		}

		// Check Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			m.writeUnauthorizedResponse(w, "Invalid authorization header format")
			return
		}

		token := parts[1]
		if token == "" {
			m.writeUnauthorizedResponse(w, "Token is required")
			return
		}

		// Generate hash of the provided token
		tokenHash := sha256.Sum256([]byte(token))
		tokenHashString := fmt.Sprintf("%x", tokenHash)

		// TODO: Extract user ID from token or use default for now
		userID := "default_user"

		// Get database instance for user
		db, err := m.dbManager.GetDB(userID)
		if err != nil {
			log.Printf("Error getting database: %v", err)
			m.writeUnauthorizedResponse(w, "Authentication failed")
			return
		}

		// Check if token exists and is valid
		var apiToken database.APIToken
		if err := db.Where("hash = ? AND deleted_at IS NULL", tokenHashString).First(&apiToken).Error; err != nil {
			log.Printf("Invalid token attempt: %v", err)
			m.writeUnauthorizedResponse(w, "Invalid token")
			return
		}

		// Update last_used timestamp
		now := time.Now()
		apiToken.LastUsed = &now
		if err := db.Save(&apiToken).Error; err != nil {
			log.Printf("Error updating token last_used: %v", err)
		}

		// Set user ID in request context for handlers to use
		// For now, we'll continue with the existing pattern
		
		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}

func (m *AuthMiddleware) writeUnauthorizedResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	response := map[string]string{
		"error":   "Unauthorized",
		"message": message,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}