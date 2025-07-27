package internal

import (
	"fmt"
	"log"
	"time"

	data_database "github.com/tkowalski/socgo/internal/data/database"
	"github.com/tkowalski/socgo/internal/process/post/data"
)

// FindPendingPostsTask finds all pending posts that are ready to be published
type FindPendingPostsTask struct{}

// NewFindPendingPostsTask creates a new find pending posts task
func NewFindPendingPostsTask() data.Task {
	return &FindPendingPostsTask{}
}

// Execute finds all pending posts that are due for publication
func (t *FindPendingPostsTask) Execute(ctx *data.PostContext) error {
	// Get database instance for user
	db, err := ctx.DB.GetDB(ctx.UserID)
	if err != nil {
		log.Printf("Error getting database: %v", err)
		return fmt.Errorf("failed to get database: %w", err)
	}

	now := time.Now()

	// Find all pending posts that are due for publication
	var pendingPosts []data_database.Post
	result := db.Where("status = ? AND (scheduled_at IS NULL OR scheduled_at <= ?)",
		data_database.PostStatusPending, now).
		Preload("Provider").
		Preload("Media").
		Find(&pendingPosts)

	if result.Error != nil {
		log.Printf("Error finding pending posts: %v", result.Error)
		return fmt.Errorf("failed to find pending posts: %w", result.Error)
	}

	log.Printf("Found %d pending posts ready for publication", len(pendingPosts))

	// Store found posts in context for processing
	ctx.PendingPosts = pendingPosts

	return nil
}
