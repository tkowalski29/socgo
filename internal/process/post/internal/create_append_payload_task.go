package internal

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/tkowalski/socgo/internal/process/post/data"
)

// CreatePayloadTask creates the JSON payload for scheduled jobs
type CreatePayloadTask struct{}

// NewCreatePayloadTask creates a new create payload task
func NewCreatePayloadTask() data.Task {
	return &CreatePayloadTask{}
}

// Execute creates the JSON payload with content and media information
func (t *CreatePayloadTask) Execute(ctx *data.PostContext) error {
	// Check if this is an append operation
	operation := ctx.Request.FormValue("operation")
	isAppend := operation == "append"

	// Create payload data with content and media information
	payloadData := struct {
		Operation      string `json:"operation,omitempty"`
		Content        string `json:"content"`
		ExistingPostID string `json:"existing_post_id,omitempty"`
		Media          []struct {
			FileName string `json:"file_name"`
			FileType string `json:"file_type"`
			FilePath string `json:"file_path"`
			FileSize int64  `json:"file_size"`
			MimeType string `json:"mime_type"`
		} `json:"media"`
	}{
		Content: ctx.Content,
		Media: make([]struct {
			FileName string `json:"file_name"`
			FileType string `json:"file_type"`
			FilePath string `json:"file_path"`
			FileSize int64  `json:"file_size"`
			MimeType string `json:"mime_type"`
		}, 0),
	}

	// Add operation-specific fields
	if isAppend {
		payloadData.Operation = "append"
		payloadData.ExistingPostID = ctx.Request.FormValue("existing_post_id")
	}

	// Add media information to payload
	for _, m := range ctx.Media {
		payloadData.Media = append(payloadData.Media, struct {
			FileName string `json:"file_name"`
			FileType string `json:"file_type"`
			FilePath string `json:"file_path"`
			FileSize int64  `json:"file_size"`
			MimeType string `json:"mime_type"`
		}{
			FileName: m.FileName,
			FileType: m.FileType,
			FilePath: m.FilePath,
			FileSize: m.FileSize,
			MimeType: m.MimeType,
		})
	}

	// Convert payload to JSON
	payloadJSON, err := json.Marshal(payloadData)
	if err != nil {
		log.Printf("Error marshaling payload data: %v", err)
		return fmt.Errorf("failed to marshal payload data: %w", err)
	}

	ctx.PayloadJSON = payloadJSON
	return nil
}
