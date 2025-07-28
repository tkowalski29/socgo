package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	data_database "github.com/tkowalski/socgo/internal/data/database"
	"github.com/tkowalski/socgo/internal/data/provider"
	"github.com/tkowalski/socgo/internal/process/post/data"
	"github.com/tkowalski/socgo/internal/service/post"
	"github.com/tkowalski/socgo/internal/social"
)

// PublishPostsTask publishes all pending posts that are ready
type PublishPostsTask struct {
	providerService *post.ProviderService
}

// NewPublishPostsTask creates a new publish posts task
func NewPublishPostsTask(providerService *post.ProviderService) data.Task {
	return &PublishPostsTask{
		providerService: providerService,
	}
}

// Execute publishes all pending posts
func (t *PublishPostsTask) Execute(ctx *data.PostContext) error {
	if len(ctx.PendingPosts) == 0 {
		log.Printf("No pending posts to publish")
		return nil
	}

	// Get database instance for user
	db, err := ctx.DB.GetDB(ctx.UserID)
	if err != nil {
		log.Printf("Error getting database: %v", err)
		return fmt.Errorf("failed to get database: %w", err)
	}

	ctx_background := context.Background()
	publishedCount := 0
	failedCount := 0

	for _, post := range ctx.PendingPosts {
		log.Printf("Publishing post ID %d for provider %s (ID: %d, Type: %s)", post.ID, post.Provider.Name, post.Provider.ID, post.Provider.Type)

		// Parse settings from database
		var settings map[string]string
		if post.Settings != "" {
			if err := json.Unmarshal([]byte(post.Settings), &settings); err != nil {
				log.Printf("Warning: Failed to parse settings for post %d: %v", post.ID, err)
				settings = nil
			} else {
				log.Printf("Loaded settings for post %d: %+v", post.ID, settings)
			}
		}

		// Convert media to provider.Media format
		var media []provider.Media
		for _, m := range post.Media {
			media = append(media, provider.Media{
				FileName: m.FileName,
				FileType: m.FileType,
				FilePath: m.FilePath,
				FileSize: m.FileSize,
				MimeType: m.MimeType,
			})
		}

		// Publish content using provider service
		postID, err := t.providerService.PublishContentByID(ctx_background, ctx.UserID, post.Provider.ID, post.Content, media, settings)
		if err != nil {
			log.Printf("Failed to publish post %d: %v", post.ID, err)

			// Update post status to failed
			post.Status = data_database.PostStatusFailed
			post.ErrorMessage = err.Error()
			post.UpdatedAt = time.Now()

			if err := db.Save(&post).Error; err != nil {
				log.Printf("Warning: Failed to update failed post %d: %v", post.ID, err)
			}

			failedCount++
			continue
		}

		// Extract external ID and URL from postID
		externalID := postID
		externalURL := ""

		log.Printf("Post %d published successfully. Provider: %s, PostID: %s", post.ID, post.Provider.Name, postID)

		// Generate external URL using helper function
		externalURL = social.BuildExternalURL(post.Provider.Type, post.Provider.Name, postID)

		// Update post with success information
		now := time.Now()
		post.ExternalID = externalID
		post.ExternalURL = externalURL
		post.PublishedAt = &now
		post.Status = data_database.PostStatusPublished
		post.ErrorMessage = ""
		post.UpdatedAt = time.Now()

		if err := db.Save(&post).Error; err != nil {
			log.Printf("Warning: Failed to update published post %d: %v", post.ID, err)
		}

		publishedCount++
	}

	log.Printf("Publishing completed: %d published, %d failed", publishedCount, failedCount)
	return nil
}
