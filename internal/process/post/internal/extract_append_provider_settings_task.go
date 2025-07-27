package internal

import (
	"strings"

	"github.com/tkowalski/socgo/internal/process/post/data"
)

// ExtractProviderSettingsTask extracts provider-specific settings from form
type ExtractProviderSettingsTask struct{}

// NewExtractProviderSettingsTask creates a new extract provider settings task
func NewExtractProviderSettingsTask() data.Task {
	return &ExtractProviderSettingsTask{}
}

// Execute extracts provider settings from form data
func (t *ExtractProviderSettingsTask) Execute(ctx *data.PostContext) error {
	settings := make(map[string]string)
	for key, values := range ctx.Request.Form {
		if len(values) > 0 && (strings.Contains(key, "_location_") ||
			strings.Contains(key, "_visibility_") ||
			strings.Contains(key, "_comments_") ||
			strings.Contains(key, "_duets_")) {
			settings[key] = values[0]
		}
	}

	ctx.Settings = settings
	return nil
}
