package internal

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/tkowalski/socgo/internal/process/post/data"
)

// ParseDataTask parses multipart form data
type ParseDataTask struct{}

// NewParseDataTask creates a new parse data task
func NewParseDataTask() data.Task {
	return &ParseDataTask{}
}

// Execute parses the multipart form data and extracts basic fields
func (t *ParseDataTask) Execute(ctx *data.PostContext) error {
	// Parse multipart form data
	if err := ctx.Request.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		log.Printf("Error parsing multipart form: %v", err)
		return fmt.Errorf("failed to parse form data: %w", err)
	}

	// Check if this is an append operation
	operation := ctx.Request.FormValue("operation")
	isAppend := operation == "append"

	// Extract existing post ID for append (only if append operation)
	if isAppend {
		existingPostIDStr := ctx.Request.FormValue("existing_post_id")
		if existingPostIDStr == "" {
			return fmt.Errorf("existing_post_id is required for append operation")
		}

		_, err := strconv.ParseUint(existingPostIDStr, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid existing_post_id: %s", existingPostIDStr)
		}
	}

	// Extract providers - handle both comma-separated and multiple fields
	var providerIDList []uint

	// Check if we have multiple provider fields
	if providers, exists := ctx.Request.Form["providers"]; exists && len(providers) > 0 {
		log.Printf("Found multiple provider fields: %v", providers)
		for _, providerIDStr := range providers {
			providerID, err := strconv.ParseUint(strings.TrimSpace(providerIDStr), 10, 32)
			if err != nil {
				return fmt.Errorf("invalid provider ID: %s", providerIDStr)
			}
			providerIDList = append(providerIDList, uint(providerID))
		}
	} else {
		// Fallback to comma-separated string (from JavaScript)
		providersStr := ctx.Request.FormValue("providers")
		log.Printf("Parsing providers - FormValue: '%s'", providersStr)

		if providersStr != "" {
			// Parse providers (comma-separated list)
			providerIDs := strings.Split(providersStr, ",")
			log.Printf("Parsing comma-separated providers: %v", providerIDs)
			for _, providerIDStr := range providerIDs {
				providerID, err := strconv.ParseUint(strings.TrimSpace(providerIDStr), 10, 32)
				if err != nil {
					return fmt.Errorf("invalid provider ID: %s", providerIDStr)
				}
				providerIDList = append(providerIDList, uint(providerID))
			}
		} else {
			log.Printf("No providers found in form data")
		}
	}

	log.Printf("Final provider IDs: %v", providerIDList)

	if len(providerIDList) == 0 {
		return fmt.Errorf("at least one provider must be selected")
	}

	ctx.ProviderIDs = providerIDList

	// Extract content based on operation type
	var content string
	if isAppend {
		content = ctx.Request.FormValue("append_content")
		if strings.TrimSpace(content) == "" {
			return fmt.Errorf("append_content is required for append operation")
		}
	} else {
		content = ctx.Request.FormValue("content")
		if strings.TrimSpace(content) == "" {
			return fmt.Errorf("content is required")
		}
	}
	ctx.Content = content

	// Extract schedule_at
	scheduleAt := ctx.Request.FormValue("schedule_at_native")
	if scheduleAt == "" {
		scheduleAt = "now"
	}
	ctx.ScheduleAt = scheduleAt

	return nil
}
