package internal

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	provider_pkg "github.com/tkowalski/socgo/internal/data/provider"
	"github.com/tkowalski/socgo/internal/process/post/data"
)

// HandleFileUploadsTask handles file uploads
type HandleFileUploadsTask struct{}

// NewHandleFileUploadsTask creates a new handle file uploads task
func NewHandleFileUploadsTask() data.Task {
	return &HandleFileUploadsTask{}
}

// Execute handles file uploads and saves them to disk
func (t *HandleFileUploadsTask) Execute(ctx *data.PostContext) error {
	var media []provider_pkg.Media

	// Check if this is an append operation
	operation := ctx.Request.FormValue("operation")
	isAppend := operation == "append"

	// Get files based on operation type
	var files []*multipart.FileHeader
	if isAppend {
		files = ctx.Request.MultipartForm.File["append_media"]
	} else {
		files = ctx.Request.MultipartForm.File["media"]
	}

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("Error opening uploaded file: %v", err)
			continue
		}
		defer file.Close()

		// Create uploads directory if it doesn't exist
		uploadDir := "uploads"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			log.Printf("Error creating uploads directory: %v", err)
			continue
		}

		// Generate unique filename
		ext := filepath.Ext(fileHeader.Filename)
		prefix := "post"
		if isAppend {
			prefix = "append"
		}
		filename := fmt.Sprintf("%s_%d_%d%s", prefix, time.Now().UnixNano(), rand.Intn(1000), ext)
		filePath := filepath.Join(uploadDir, filename)

		// Create the file
		dst, err := os.Create(filePath)
		if err != nil {
			log.Printf("Error creating file: %v", err)
			continue
		}
		defer dst.Close()

		// Copy file content
		if _, err := io.Copy(dst, file); err != nil {
			log.Printf("Error copying file: %v", err)
			continue
		}

		// Add to media list
		fileType := "image" // Default to image
		mimeType := fileHeader.Header.Get("Content-Type")
		if strings.Contains(mimeType, "video/") {
			fileType = "video"
		}

		media = append(media, provider_pkg.Media{
			FileName: fileHeader.Filename,
			FilePath: filePath,
			FileType: fileType,
			FileSize: fileHeader.Size,
			MimeType: mimeType,
		})
	}

	ctx.Media = media
	return nil
}
