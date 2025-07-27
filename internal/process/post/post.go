package post

import (
	"fmt"
	"log"

	"github.com/tkowalski/socgo/internal/process/post/data"
	"github.com/tkowalski/socgo/internal/process/post/internal"
	"github.com/tkowalski/socgo/internal/service/post"
)

// PostProcess orchestrates the regular post creation process
type PostProcess interface {
	Execute(ctx *data.PostContext) error
}

// PostProcessImpl implements PostProcess
type PostProcessImpl struct {
	tasks []data.Task
}

// NewPostProcess creates a new regular post process with all required tasks
func NewPostProcess(providerService *post.ProviderService) PostProcess {
	return &PostProcessImpl{
		tasks: []data.Task{
			internal.NewValidateRequestTask(),
			internal.NewParseDataTask(),
			internal.NewValidateContentTask(),
			internal.NewHandleFileUploadsTask(),
			internal.NewExtractProviderSettingsTask(),
			internal.NewValidateProvidersTask(providerService),
			internal.NewSavePostTask(),
		},
	}
}

// Execute runs all tasks in sequence
func (p *PostProcessImpl) Execute(ctx *data.PostContext) error {
	log.Printf("Starting regular post process for user: %s", ctx.UserID)

	for i, task := range p.tasks {
		log.Printf("Executing post task %d/%d", i+1, len(p.tasks))

		if err := task.Execute(ctx); err != nil {
			log.Printf("Post task %d failed: %v", i+1, err)
			ctx.Errors = append(ctx.Errors, err)
			return fmt.Errorf("post task %d failed: %w", i+1, err)
		}
	}

	log.Printf("Regular post process completed successfully for user: %s", ctx.UserID)
	return nil
}
