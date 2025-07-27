package internal

import (
	"fmt"
	"log"
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

	// Parse scheduled time
	var scheduledTime *time.Time
	if ctx.ScheduleAt == "now" {
		now := time.Now()
		scheduledTime = &now
	} else {
		parsedTime, err := time.Parse(time.RFC3339, ctx.ScheduleAt)
		if err != nil {
			return fmt.Errorf("invalid schedule_at format: %w", err)
		}
		scheduledTime = &parsedTime
	}

	// Create posts for each provider
	var createdPosts []data_database.Post
	for _, providerID := range ctx.ProviderIDs {
		// Create post record
		post := data_database.Post{
			Content:     ctx.Content,
			UserID:      ctx.UserID,
			ProviderID:  providerID,
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
