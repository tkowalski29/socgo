package post

import (
	"fmt"

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
	for i, task := range p.tasks {
		if err := task.Execute(ctx); err != nil {
			ctx.Errors = append(ctx.Errors, err)
			return fmt.Errorf("scheduler task %d failed: %w", i+1, err)
		}
	}
	return nil
}
