package internal

import (
	"fmt"
	"strings"

	"github.com/tkowalski/socgo/internal/process/post/data"
)

// ValidateContentTask validates the content and schedule
type ValidateContentTask struct{}

// NewValidateContentTask creates a new validate content task
func NewValidateContentTask() data.Task {
	return &ValidateContentTask{}
}

// Execute validates content and schedule
func (t *ValidateContentTask) Execute(ctx *data.PostContext) error {
	// Validate content
	if len(strings.TrimSpace(ctx.Content)) == 0 {
		return fmt.Errorf("content cannot be empty")
	}

	// Check if this is an append operation
	operation := ctx.Request.FormValue("operation")
	isAppend := operation == "append"

	// Additional validation for content based on operation type
	if isAppend {
		if len(ctx.Content) > 10000 { // Max 10k characters for append
			return fmt.Errorf("append content too long, maximum 10000 characters allowed")
		}
	} else {
		if len(ctx.Content) > 2200 { // Max 2200 characters for regular posts
			return fmt.Errorf("post content too long, maximum 2200 characters allowed")
		}
	}

	// Validate schedule format if not "now"
	if ctx.ScheduleAt != "now" {
		if _, err := ParseScheduleTime(ctx.ScheduleAt); err != nil {
			return fmt.Errorf("invalid schedule_at format, use ISO8601 or 'now': %w", err)
		}
	}

	return nil
}
