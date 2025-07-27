package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	data_database "github.com/tkowalski/socgo/internal/data/database"
	"github.com/tkowalski/socgo/internal/data/provider"
	"github.com/tkowalski/socgo/internal/service/database"
	"github.com/tkowalski/socgo/internal/service/notifications"
	"github.com/tkowalski/socgo/internal/service/post"
	"gorm.io/gorm"
)

// Scheduler manages scheduled jobs execution
type Scheduler struct {
	dbManager           *database.Manager
	providerService     *post.ProviderService
	notificationService *notifications.Service
	ticker              *time.Ticker
	stopChan            chan struct{}
}

// New creates a new scheduler instance
func New(dbManager *database.Manager, providerService *post.ProviderService, notificationService *notifications.Service) *Scheduler {
	return &Scheduler{
		dbManager:           dbManager,
		providerService:     providerService,
		notificationService: notificationService,
		stopChan:            make(chan struct{}),
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
	var jobs []data_database.ScheduledJob
	now := time.Now()

	result := db.Where("status = ? AND scheduled_at <= ?", data_database.JobStatusPending, now).
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
func (s *Scheduler) processJob(ctx context.Context, userID string, db *gorm.DB, job *data_database.ScheduledJob) error {
	log.Printf("Processing job %d: %s for user %s", job.ID, job.JobType, userID)

	// Mark job as executing
	job.Status = data_database.JobStatusExecuting
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
func (s *Scheduler) processPublishPostJob(ctx context.Context, userID string, db *gorm.DB, job *data_database.ScheduledJob) error {
	// Get provider name from the job
	if job.Provider.Name == "" {
		return s.markJobFailed(db, job, "Provider name not found")
	}

	// Parse PayloadData to extract content and media information
	var payloadData struct {
		Content string `json:"content"`
		Media   []struct {
			FileName string `json:"file_name"`
			FileType string `json:"file_type"`
			FilePath string `json:"file_path"`
			FileSize int64  `json:"file_size"`
			MimeType string `json:"mime_type"`
		} `json:"media"`
	}

	// Try to parse as JSON first, if it fails, treat as plain text
	if err := json.Unmarshal([]byte(job.PayloadData), &payloadData); err != nil {
		// If it's not JSON, treat as plain text content
		payloadData.Content = job.PayloadData
	}

	// Convert media information to provider.Media format
	var media []provider.Media
	for _, m := range payloadData.Media {
		media = append(media, provider.Media{
			FileName: m.FileName,
			FileType: m.FileType,
			FilePath: m.FilePath,
			FileSize: m.FileSize,
			MimeType: m.MimeType,
		})
	}

	// Publish content using provider service
	postID, err := s.providerService.PublishContent(ctx, userID, job.Provider.Name, payloadData.Content, media, nil)
	if err != nil {
		// Save failed post with error information
		post := data_database.Post{
			Content:      payloadData.Content,
			UserID:       userID,
			ProviderID:   job.ProviderID,
			Status:       data_database.PostStatusFailed,
			ErrorMessage: err.Error(),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := db.Create(&post).Error; err != nil {
			log.Printf("Warning: Failed to save failed post record for job %d: %v", job.ID, err)
		}

		// Create notification for failed scheduled post
		if s.notificationService != nil {
			s.notificationService.CreateNotification(
				userID,
				"schedule",
				"error",
				"Błąd publikacji zaplanowanego posta",
				fmt.Sprintf("Nie udało się opublikować posta na %s: %s", job.Provider.Name, err.Error()),
				nil,
			)
		}

		return s.markJobFailed(db, job, "Failed to publish content: "+err.Error())
	}

	// Extract external ID and URL from postID
	externalID := postID
	externalURL := ""

	log.Printf("Job %d: Provider Type: %s, Provider Name: %s, PostID: %s", job.ID, job.Provider.Type, job.Provider.Name, postID)

	// For Facebook, postID jest postem w formacie {pageId_postId}
	if job.Provider.Type == "facebook" && postID != "" {
		externalURL = fmt.Sprintf("https://www.facebook.com/%s", postID)
		log.Printf("Job %d: Generated Facebook URL: %s", job.ID, externalURL)
	}
	// Instagram: postID to media ID, można wygenerować link jeśli znamy userID
	if job.Provider.Type == "instagram" && postID != "" {
		// Instagram nie udostępnia bezpośredniego linku po API, ale można spróbować:
		externalURL = fmt.Sprintf("https://www.instagram.com/p/%s/", postID)
		log.Printf("Job %d: Generated Instagram URL: %s", job.ID, externalURL)
	}
	// TikTok: postID to video ID
	if job.Provider.Type == "tiktok" && postID != "" {
		externalURL = fmt.Sprintf("https://www.tiktok.com/@%s/video/%s", job.Provider.Name, postID)
		log.Printf("Job %d: Generated TikTok URL: %s", job.ID, externalURL)
	}

	if externalURL == "" {
		log.Printf("Job %d: No external URL generated for provider type: %s", job.ID, job.Provider.Type)
	}

	// Set published time and status
	now := time.Now()
	status := data_database.PostStatusPublished

	// Create post record
	post := data_database.Post{
		Content:      payloadData.Content,
		UserID:       userID,
		ProviderID:   job.ProviderID,
		ExternalID:   externalID,
		ExternalURL:  externalURL,
		PublishedAt:  &now,
		Status:       status,
		ErrorMessage: "",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := db.Create(&post).Error; err != nil {
		log.Printf("Warning: Failed to save post record for job %d: %v", job.ID, err)
		// Continue - post was published successfully
	}

	// Save media files to database if they exist
	if len(media) > 0 {
		log.Printf("Saving %d media files to database for job %d", len(media), job.ID)
		for _, m := range media {
			mediaRecord := data_database.Media{
				PostID:    post.ID,
				FileName:  m.FileName,
				FilePath:  m.FilePath,
				FileType:  m.FileType,
				FileSize:  m.FileSize,
				MimeType:  m.MimeType,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			if err := db.Create(&mediaRecord).Error; err != nil {
				log.Printf("Error saving media file %s for job %d: %v", m.FileName, job.ID, err)
				// Continue with other media files
			} else {
				log.Printf("Media file saved for job %d: %s", job.ID, m.FileName)
			}
		}
	}

	// Mark job as completed
	job.Status = data_database.JobStatusCompleted
	job.ExecutedAt = &[]time.Time{time.Now()}[0]
	job.UpdatedAt = time.Now()

	if err := db.Save(job).Error; err != nil {
		return err
	}

	// Create notification for successful scheduled post
	if s.notificationService != nil {
		postID := uint(post.ID)
		s.notificationService.CreateNotification(
			userID,
			"schedule",
			"success",
			"Zaplanowany post opublikowany",
			fmt.Sprintf("Post został opublikowany na %s o %s", job.Provider.Name, now.Format("02.01.2006 15:04")),
			&postID,
		)
	}

	log.Printf("Job %d completed successfully. Post ID: %s", job.ID, postID)
	return nil
}

// markJobFailed marks a job as failed with error message
func (s *Scheduler) markJobFailed(db *gorm.DB, job *data_database.ScheduledJob, errorMsg string) error {
	job.Status = data_database.JobStatusFailed
	job.ErrorMsg = errorMsg
	job.UpdatedAt = time.Now()

	if err := db.Save(job).Error; err != nil {
		return err
	}

	log.Printf("Job %d failed: %s", job.ID, errorMsg)
	return nil
}
