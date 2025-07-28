package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	data_database "github.com/tkowalski/socgo/internal/data/database"
	"github.com/tkowalski/socgo/internal/process/post/data"
)

// SavePostTask saves the post to the database with pending status
type SavePostTask struct{}

// NewSavePostTask creates a new save post task
func NewSavePostTask() data.Task {
	return &SavePostTask{}
}

// Execute saves the post to the database
func (t *SavePostTask) Execute(ctx *data.PostContext) error {
	// Get database instance for user
	db, err := ctx.DB.GetDB(ctx.UserID)
	if err != nil {
		log.Printf("Error getting database: %v", err)
		return fmt.Errorf("failed to get database: %w", err)
	}

	// Parse scheduled time using flexible parser
	scheduledTime, err := ParseScheduleTime(ctx.ScheduleAt)
	if err != nil {
		return fmt.Errorf("invalid schedule_at format: %w", err)
	}

	// Create posts for each provider
	var createdPosts []data_database.Post
	for _, providerID := range ctx.ProviderIDs {
		// Get provider details to determine type
		var provider data_database.Provider
		if err := db.First(&provider, providerID).Error; err != nil {
			log.Printf("Error getting provider %d: %v", providerID, err)
			return fmt.Errorf("failed to get provider %d: %w", providerID, err)
		}

		// Filter settings for this specific provider
		providerSettings := make(map[string]string)
		for key, value := range ctx.Settings {
			// Check if setting belongs to this provider (e.g., facebook_location_Test1)
			if strings.HasPrefix(key, provider.Type+"_") && strings.HasSuffix(key, "_"+provider.Name) {
				// Extract the setting name without prefix and suffix
				settingName := strings.TrimPrefix(key, provider.Type+"_")
				settingName = strings.TrimSuffix(settingName, "_"+provider.Name)
				providerSettings[settingName] = value
			}
		}

		// Convert settings to JSON
		settingsJSON := ""
		if len(providerSettings) > 0 {
			settingsBytes, err := json.Marshal(providerSettings)
			if err != nil {
				log.Printf("Error marshaling settings for provider %d: %v", providerID, err)
				return fmt.Errorf("failed to marshal settings: %w", err)
			}
			settingsJSON = string(settingsBytes)
			log.Printf("Saving settings for provider %d (%s): %s", providerID, provider.Type, settingsJSON)
		}

		// Create post record
		post := data_database.Post{
			Content:     ctx.Content,
			UserID:      ctx.UserID,
			ProviderID:  providerID,
			Settings:    settingsJSON,
			ScheduledAt: scheduledTime,
			Status:      data_database.PostStatusPending,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Save post to database
		if err := db.Create(&post).Error; err != nil {
			log.Printf("Error creating post for provider %d: %v", providerID, err)
			return fmt.Errorf("failed to create post for provider %d: %w", providerID, err)
		}

		// Save media files associated with this post
		for _, media := range ctx.Media {
			mediaRecord := data_database.Media{
				PostID:    post.ID,
				FileName:  media.FileName,
				FilePath:  media.FilePath,
				FileType:  media.FileType,
				FileSize:  media.FileSize,
				MimeType:  media.MimeType,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			if err := db.Create(&mediaRecord).Error; err != nil {
				log.Printf("Error creating media record for post %d: %v", post.ID, err)
				// Continue with other media files even if one fails
				continue
			}
		}

		createdPosts = append(createdPosts, post)
		log.Printf("Created post with ID %d for provider %d", post.ID, providerID)
	}

	log.Printf("Successfully created %d posts in database", len(createdPosts))
	return nil
}
