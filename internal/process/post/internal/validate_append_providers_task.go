package internal

import (
	"fmt"
	"log"

	data_database "github.com/tkowalski/socgo/internal/data/database"
	"github.com/tkowalski/socgo/internal/process/post/data"
	"github.com/tkowalski/socgo/internal/service/post"
)

// ValidateProvidersTask validates that all providers exist and are configured
type ValidateProvidersTask struct {
	providerService *post.ProviderService
}

// NewValidateProvidersTask creates a new validate providers task
func NewValidateProvidersTask(providerService *post.ProviderService) data.Task {
	return &ValidateProvidersTask{
		providerService: providerService,
	}
}

// Execute validates that all providers exist and are configured
func (t *ValidateProvidersTask) Execute(ctx *data.PostContext) error {
	// Get database instance for user
	db, err := ctx.DB.GetDB(ctx.UserID)
	if err != nil {
		log.Printf("Error getting database: %v", err)
		return fmt.Errorf("failed to get database: %w", err)
	}

	// Check if this is an append operation
	operation := ctx.Request.FormValue("operation")
	isAppend := operation == "append"

	// Check if all providers are configured
	for _, providerID := range ctx.ProviderIDs {
		var provider data_database.Provider
		if err := db.First(&provider, providerID).Error; err != nil {
			return fmt.Errorf("provider not found with ID %d: %w", providerID, err)
		}

		// Check if provider is configured
		isConfigured, err := t.providerService.IsProviderConfigured(ctx.UserID, provider.Name)
		if err != nil {
			log.Printf("Error checking provider configuration: %v", err)
			return fmt.Errorf("failed to check provider configuration: %w", err)
		}
		if !isConfigured {
			return fmt.Errorf("provider %s is not configured", provider.Name)
		}

		// Additional validation for append-specific requirements
		if isAppend {
			if !t.supportsAppend(provider.Type) {
				return fmt.Errorf("provider %s does not support append operations", provider.Name)
			}
		}
	}

	return nil
}

// supportsAppend checks if the provider type supports append operations
func (t *ValidateProvidersTask) supportsAppend(providerType string) bool {
	// Define which provider types support append operations
	supportedTypes := map[string]bool{
		"facebook":  true,
		"instagram": true,
		"tiktok":    false, // Example: TikTok might not support append
	}

	supported, exists := supportedTypes[providerType]
	return exists && supported
}
