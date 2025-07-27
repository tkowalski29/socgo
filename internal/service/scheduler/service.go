package scheduler

import (
	"log"
	"time"

	data_database "github.com/tkowalski/socgo/internal/data/database"
	provider_pkg "github.com/tkowalski/socgo/internal/data/provider"
	process_post "github.com/tkowalski/socgo/internal/process/post"
	process_post_data "github.com/tkowalski/socgo/internal/process/post/data"
	"github.com/tkowalski/socgo/internal/service/database"
	"github.com/tkowalski/socgo/internal/service/notifications"
	post_service "github.com/tkowalski/socgo/internal/service/post"
)

// Scheduler manages post publishing execution
type Service struct {
	dbManager           *database.Manager
	providerService     *post_service.ProviderService
	notificationService *notifications.Service
	ticker              *time.Ticker
	stopChan            chan struct{}
}

// New creates a new scheduler instance
func New(dbManager *database.Manager, providerService *post_service.ProviderService, notificationService *notifications.Service) *Service {
	return &Service{
		dbManager:           dbManager,
		providerService:     providerService,
		notificationService: notificationService,
		stopChan:            make(chan struct{}),
	}
}

// Start begins the scheduler worker
func (s *Service) Start() {
	log.Println("Starting post scheduler...")

	// Run immediately on start
	s.processPosts()

	// Schedule to run every minute
	s.ticker = time.NewTicker(1 * time.Minute)

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.processPosts()
			case <-s.stopChan:
				return
			}
		}
	}()
}

// Stop gracefully stops the scheduler
func (s *Service) Stop() {
	log.Println("Stopping post scheduler...")

	if s.ticker != nil {
		s.ticker.Stop()
	}

	close(s.stopChan)
}

// processPosts processes all pending posts that are ready to be published
func (s *Service) processPosts() {
	log.Println("Processing pending posts...")

	// Get all user databases
	userDatabases := s.dbManager.GetAllUserDatabases()

	for userID := range userDatabases {
		if err := s.processUserPosts(userID); err != nil {
			log.Printf("Error processing posts for user %s: %v", userID, err)
		}
	}
}

// processUserPosts processes posts for a specific user
func (s *Service) processUserPosts(userID string) error {
	// Create scheduler process
	schedulerProcess := process_post.NewSchedulerProcess(s.providerService)

	// Create post context for scheduler
	postCtx := &process_post_data.PostContext{
		UserID:       userID,
		DB:           s.dbManager,
		Content:      "",
		ScheduleAt:   "",
		Media:        []provider_pkg.Media{},
		Settings:     make(map[string]string),
		ProviderIDs:  []uint{},
		PayloadJSON:  []byte{},
		Errors:       []error{},
		PendingPosts: []data_database.Post{},
	}

	// Execute the scheduler process
	if err := schedulerProcess.Execute(postCtx); err != nil {
		log.Printf("Scheduler process failed for user %s: %v", userID, err)
		return err
	}

	return nil
}
