package post

import (
	"fmt"
	"log"

	"github.com/tkowalski/socgo/internal/process/post/data"
	"github.com/tkowalski/socgo/internal/process/post/internal"
	"github.com/tkowalski/socgo/internal/service/post"
)

// SchedulerProcess orchestrates the post publishing process for scheduler
type SchedulerProcess interface {
	Execute(ctx *data.PostContext) error
}

// SchedulerProcessImpl implements SchedulerProcess
type SchedulerProcessImpl struct {
	tasks []data.Task
}

// NewSchedulerProcess creates a new scheduler process with all required tasks
func NewSchedulerProcess(providerService *post.ProviderService) SchedulerProcess {
	return &SchedulerProcessImpl{
		tasks: []data.Task{
			internal.NewFindPendingPostsTask(),
			internal.NewPublishPostsTask(providerService),
		},
	}
}

// Execute runs all tasks in sequence
func (p *SchedulerProcessImpl) Execute(ctx *data.PostContext) error {
	log.Printf("Starting scheduler process for user: %s", ctx.UserID)

	for i, task := range p.tasks {
		log.Printf("Executing scheduler task %d/%d", i+1, len(p.tasks))

		if err := task.Execute(ctx); err != nil {
			log.Printf("Scheduler task %d failed: %v", i+1, err)
			ctx.Errors = append(ctx.Errors, err)
			return fmt.Errorf("scheduler task %d failed: %w", i+1, err)
		}
	}

	log.Printf("Scheduler process completed successfully for user: %s", ctx.UserID)
	return nil
}
