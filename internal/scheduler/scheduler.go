package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/tkowalski/socgo/internal/database"
	"github.com/tkowalski/socgo/internal/providers"
	"gorm.io/gorm"
)

// Scheduler manages scheduled jobs execution
type Scheduler struct {
	dbManager       *database.Manager
	providerService *providers.ProviderService
	ticker          *time.Ticker
	stopChan        chan struct{}
}

// New creates a new scheduler instance
func New(dbManager *database.Manager, providerService *providers.ProviderService) *Scheduler {
	return &Scheduler{
		dbManager:       dbManager,
		providerService: providerService,
		stopChan:        make(chan struct{}),
	}
}

// Start begins the scheduler worker
func (s *Scheduler) Start() {
	log.Println("Starting job scheduler...")

	// Run immediately on start
	s.processJobs()

	// Schedule to run every minute
	s.ticker = time.NewTicker(1 * time.Minute)

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.processJobs()
			case <-s.stopChan:
				return
			}
		}
	}()
}

// Stop gracefully stops the scheduler
func (s *Scheduler) Stop() {
	log.Println("Stopping job scheduler...")

	if s.ticker != nil {
		s.ticker.Stop()
	}

	close(s.stopChan)
}

// processJobs processes all pending scheduled jobs
func (s *Scheduler) processJobs() {
	log.Println("Processing scheduled jobs...")

	ctx := context.Background()

	// Get all user databases
	userDatabases := s.dbManager.GetAllUserDatabases()

	for userID, db := range userDatabases {
		if err := s.processUserJobs(ctx, userID, db); err != nil {
			log.Printf("Error processing jobs for user %s: %v", userID, err)
		}
	}
}

// processUserJobs processes jobs for a specific user
func (s *Scheduler) processUserJobs(ctx context.Context, userID string, db *gorm.DB) error {
	// Get pending jobs that are due
	var jobs []database.ScheduledJob
	now := time.Now()

	result := db.Where("status = ? AND scheduled_at <= ?", database.JobStatusPending, now).
		Preload("Provider").
		Find(&jobs)

	if result.Error != nil {
		return result.Error
	}

	log.Printf("Found %d pending jobs for user %s", len(jobs), userID)

	// Process each job
	for _, job := range jobs {
		if err := s.processJob(ctx, userID, db, &job); err != nil {
			log.Printf("Error processing job %d: %v", job.ID, err)
		}
	}

	return nil
}

// processJob processes a single scheduled job
func (s *Scheduler) processJob(ctx context.Context, userID string, db *gorm.DB, job *database.ScheduledJob) error {
	log.Printf("Processing job %d: %s for user %s", job.ID, job.JobType, userID)

	// Mark job as executing
	job.Status = database.JobStatusExecuting
	job.UpdatedAt = time.Now()

	if err := db.Save(job).Error; err != nil {
		return err
	}

	// Process different job types
	switch job.JobType {
	case "publish_post":
		return s.processPublishPostJob(ctx, userID, db, job)
	default:
		return s.markJobFailed(db, job, "Unknown job type: "+job.JobType)
	}
}

// processPublishPostJob processes a publish post job
func (s *Scheduler) processPublishPostJob(ctx context.Context, userID string, db *gorm.DB, job *database.ScheduledJob) error {
	// Get provider name from the job
	if job.Provider.Name == "" {
		return s.markJobFailed(db, job, "Provider name not found")
	}

	// Publish content using provider service
	postID, err := s.providerService.PublishContent(ctx, userID, job.Provider.Name, job.PayloadData)
	if err != nil {
		return s.markJobFailed(db, job, "Failed to publish content: "+err.Error())
	}

	// Create post record
	post := database.Post{
		Content:    job.PayloadData,
		UserID:     userID,
		ProviderID: job.ProviderID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := db.Create(&post).Error; err != nil {
		log.Printf("Warning: Failed to save post record for job %d: %v", job.ID, err)
		// Continue - post was published successfully
	}

	// Mark job as completed
	job.Status = database.JobStatusCompleted
	job.ExecutedAt = &[]time.Time{time.Now()}[0]
	job.UpdatedAt = time.Now()

	if err := db.Save(job).Error; err != nil {
		return err
	}

	log.Printf("Job %d completed successfully. Post ID: %s", job.ID, postID)
	return nil
}

// markJobFailed marks a job as failed with error message
func (s *Scheduler) markJobFailed(db *gorm.DB, job *database.ScheduledJob, errorMsg string) error {
	job.Status = database.JobStatusFailed
	job.ErrorMsg = errorMsg
	job.UpdatedAt = time.Now()

	if err := db.Save(job).Error; err != nil {
		return err
	}

	log.Printf("Job %d failed: %s", job.ID, errorMsg)
	return nil
}
