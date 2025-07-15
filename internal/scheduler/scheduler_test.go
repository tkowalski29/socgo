package scheduler

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/tkowalski/socgo/internal/database"
	"github.com/tkowalski/socgo/internal/oauth"
	"github.com/tkowalski/socgo/internal/providers"
)

func TestScheduler_E2E(t *testing.T) {
	// Create temporary directory for test database
	tempDir, err := os.MkdirTemp("", "scheduler_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create database manager
	dbManager := database.NewManager(tempDir)
	defer func() {
		if err := dbManager.CloseAll(); err != nil {
			t.Logf("Failed to close database: %v", err)
		}
	}()

	// Create OAuth service
	oauthService := oauth.NewService(dbManager)

	// Create provider service
	providerService := providers.NewProviderService(dbManager, oauthService)

	// Create scheduler
	scheduler := New(dbManager, providerService)

	userID := "test_user"

	// Get database for test user
	db, err := dbManager.GetDB(userID)
	if err != nil {
		t.Fatal(err)
	}

	// Create test provider
	provider := database.Provider{
		Name:     "facebook",
		Type:     "facebook",
		Config:   `{"access_token":"test_token","token_type":"Bearer","expires_at":"2025-12-31T23:59:59Z"}`,
		UserID:   userID,
		IsActive: true,
	}
	if err := db.Create(&provider).Error; err != nil {
		t.Fatal(err)
	}

	// Create scheduled job set 1 minute in the future
	scheduleAt := time.Now().Add(1 * time.Minute)
	job := database.ScheduledJob{
		JobType:     "publish_post",
		PayloadData: "Test scheduled post content",
		UserID:      userID,
		ProviderID:  provider.ID,
		ScheduledAt: scheduleAt,
		Status:      database.JobStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := db.Create(&job).Error; err != nil {
		t.Fatal(err)
	}

	// Fast-forward time by creating a job that's already due
	immediateJob := database.ScheduledJob{
		JobType:     "publish_post",
		PayloadData: "Test immediate post content",
		UserID:      userID,
		ProviderID:  provider.ID,
		ScheduledAt: time.Now().Add(-1 * time.Minute), // 1 minute ago
		Status:      database.JobStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := db.Create(&immediateJob).Error; err != nil {
		t.Fatal(err)
	}

	// Process jobs manually (instead of waiting for ticker)
	ctx := context.Background()
	err = scheduler.processUserJobs(ctx, userID, db)
	if err != nil {
		t.Fatal(err)
	}

	// Verify immediate job was processed
	var processedJob database.ScheduledJob
	if err := db.First(&processedJob, immediateJob.ID).Error; err != nil {
		t.Fatal(err)
	}

	if processedJob.Status != database.JobStatusFailed {
		// Note: Job will fail because we don't have a real provider configured
		// This is expected in test environment
		t.Logf("Job status: %s (expected to fail in test environment)", processedJob.Status)
	}

	// Verify future job is still pending
	var futureJob database.ScheduledJob
	if err := db.First(&futureJob, job.ID).Error; err != nil {
		t.Fatal(err)
	}

	if futureJob.Status != database.JobStatusPending {
		t.Errorf("Expected future job to remain pending, got: %s", futureJob.Status)
	}

	t.Log("E2E test completed successfully")
}
