package internal

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/tkowalski/socgo/internal/process/post/data"
)

// ValidateRequestTask validates the incoming request
type ValidateRequestTask struct{}

// NewValidateRequestTask creates a new validate request task
func NewValidateRequestTask() data.Task {
	return &ValidateRequestTask{}
}

// Execute validates the request method and content type
func (t *ValidateRequestTask) Execute(ctx *data.PostContext) error {
	if ctx.Request.Method != http.MethodPost {
		return fmt.Errorf("method not allowed: %s", ctx.Request.Method)
	}

	contentType := ctx.Request.Header.Get("Content-Type")
	if !strings.Contains(contentType, "multipart/form-data") {
		return fmt.Errorf("only multipart form data is supported, got: %s", contentType)
	}

	return nil
}
