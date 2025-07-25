package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	data_database "github.com/tkowalski/socgo/internal/data/database"
	provider_pkg "github.com/tkowalski/socgo/internal/data/provider"
	"github.com/tkowalski/socgo/internal/service/database"
	"github.com/tkowalski/socgo/internal/service/post"
	"github.com/tkowalski/socgo/web/templates"
	"github.com/tkowalski/socgo/web/templates/component"
	"github.com/tkowalski/socgo/web/templates/social"
)

// WebHandler handles web page requests
type WebHandler struct {
	dbManager       *database.Manager
	providerService *post.ProviderService
	cache           *ComponentCache
}

// PageData holds common data for all pages
type PageData struct {
	Title        string
	CurrentPage  string
	FlashMessage string
	FlashType    string
	Data         interface{}
}

// NewWebHandler creates a new WebHandler instance
func NewWebHandler(dbManager *database.Manager, providerService *post.ProviderService) *WebHandler {
	return &WebHandler{
		dbManager:       dbManager,
		providerService: providerService,
		cache:           NewComponentCache(),
	}
}

// Helper method to send flash messages via HTMX headers
func (h *WebHandler) setFlashMessage(w http.ResponseWriter, message, flashType string) {
	w.Header().Set("HX-Trigger", "flash-message")
	w.Header().Set("HX-Flash-Message", message)
	w.Header().Set("HX-Flash-Type", flashType)
}

// Helper method to redirect with flash message
func (h *WebHandler) redirectWithFlash(w http.ResponseWriter, r *http.Request, url, message, flashType string) {
	flashParam := "&flash=" + strings.ReplaceAll(message, " ", "+")
	if flashType != "" {
		flashParam += "&flash_type=" + flashType
	}

	connector := "?"
	if strings.Contains(url, "?") {
		connector = "&"
	}

	http.Redirect(w, r, url+connector+flashParam[1:], http.StatusSeeOther)
}

// CalendarPage handles the calendar page
func (h *WebHandler) CalendarPage(w http.ResponseWriter, r *http.Request) {
	// Get flash message from query parameters
	flashMessage := ""
	flashType := "info"
	if flashMsg := r.URL.Query().Get("flash"); flashMsg != "" {
		if decoded, err := url.QueryUnescape(flashMsg); err == nil {
			flashMessage = decoded
		} else {
			flashMessage = flashMsg
		}
		flashType = r.URL.Query().Get("flash_type")
		if flashType == "" {
			flashType = "info"
		}
	}

	// Create layout data
	layoutData := templates.LayoutData{
		Title:        "Calendar",
		CurrentPage:  "calendar",
		FlashMessage: flashMessage,
		FlashType:    flashType,
		Content:      templates.CalendarContent(),
	}

	// Render the layout
	w.Header().Set("Content-Type", "text/html")
	layoutComponent := templates.Layout(layoutData)
	if err := layoutComponent.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering calendar page: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// HandlePost handles form submissions for creating posts with multiple providers and media files
func (h *WebHandler) HandlePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.setFlashMessage(w, "Method not allowed", "error")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form data
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		h.setFlashMessage(w, "Invalid form data", "error")
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	content := strings.TrimSpace(r.FormValue("content"))
	providers := r.Form["providers"] // Multiple providers
	publishType := r.FormValue("publish_type")
	scheduleAt := r.FormValue("schedule_at")

	// Basic validation
	if content == "" {
		h.setFlashMessage(w, "Content is required", "error")
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}
	if len(providers) == 0 {
		h.setFlashMessage(w, "Please select at least one platform", "error")
		http.Error(w, "At least one provider is required", http.StatusBadRequest)
		return
	}

	// Validate content length
	if len(content) > 2200 {
		h.setFlashMessage(w, "Content is too long (max 2200 characters)", "error")
		http.Error(w, "Content is too long", http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)
	db, err := h.dbManager.GetDB(userID)
	if err != nil {
		log.Printf("Error getting database: %v", err)
		h.setFlashMessage(w, "Internal server error", "error")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Handle media files
	var mediaFiles []data_database.Media
	if files := r.MultipartForm.File["media"]; len(files) > 0 {
		mediaFiles, err = h.handleMediaUpload(files, userID)
		if err != nil {
			log.Printf("Error handling media upload: %v", err)
			h.setFlashMessage(w, "Error uploading media files", "error")
			http.Error(w, "Error uploading media files", http.StatusInternalServerError)
			return
		}
	}

	// Determine schedule time
	var scheduledTime *time.Time
	if publishType == "scheduled" && scheduleAt != "" {
		if t, err := time.Parse("2006-01-02T15:04", scheduleAt); err == nil {
			if t.Before(time.Now()) {
				h.setFlashMessage(w, "Scheduled time must be in the future", "error")
				http.Error(w, "Scheduled time must be in the future", http.StatusBadRequest)
				return
			}
			scheduledTime = &t
		} else {
			h.setFlashMessage(w, "Invalid schedule format", "error")
			http.Error(w, "Invalid schedule format", http.StatusBadRequest)
			return
		}
	}

	// Extract provider-specific settings
	providerSettings := h.extractProviderSettings(r.Form)

	// Create posts for each selected provider
	var results []string
	var hasErrors bool

	for _, providerIDStr := range providers {
		providerID, err := strconv.ParseUint(providerIDStr, 10, 32)
		if err != nil {
			log.Printf("Invalid provider ID: %s", providerIDStr)
			continue
		}

		// Validate provider exists and is configured
		var provider data_database.Provider
		if err := db.First(&provider, providerID).Error; err != nil {
			log.Printf("Provider not found: %d", providerID)
			results = append(results, fmt.Sprintf("Provider %d: not found", providerID))
			hasErrors = true
			continue
		}

		// Check if provider is configured
		configured, err := h.providerService.IsProviderConfigured(userID, provider.Name)
		if err != nil || !configured {
			results = append(results, fmt.Sprintf("%s: not connected", provider.Name))
			hasErrors = true
			continue
		}

		// Get provider-specific settings
		settings := providerSettings[provider.Name]

		// Create post record
		post := data_database.Post{
			Content:    content,
			UserID:     userID,
			ProviderID: uint(providerID),
			Media:      mediaFiles,
			Status:     data_database.PostStatusPending, // Start with pending status
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		if scheduledTime != nil {
			// Create scheduled job
			job := data_database.ScheduledJob{
				JobType:     "publish_post",
				PayloadData: content,
				UserID:      userID,
				ProviderID:  uint(providerID),
				ScheduledAt: *scheduledTime,
				Status:      "pending",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			if err := db.Create(&job).Error; err != nil {
				log.Printf("Error creating scheduled job: %v", err)
				results = append(results, fmt.Sprintf("%s: failed to schedule", provider.Name))
				hasErrors = true
				continue
			}

			results = append(results, fmt.Sprintf("%s: scheduled for %s", provider.Name, scheduledTime.Format("2006-01-02 15:04")))
		} else {
			// Immediate publishing
			ctx := context.Background()

			// Convert database media to provider media
			var providerMedia []provider_pkg.Media
			for _, m := range mediaFiles {
				providerMedia = append(providerMedia, provider_pkg.Media{
					FileName: m.FileName,
					FilePath: m.FilePath,
					FileType: m.FileType,
					FileSize: m.FileSize,
					MimeType: m.MimeType,
				})
			}

			postID, err := h.providerService.PublishContent(ctx, userID, provider.Name, content, providerMedia, settings)
			if err != nil {
				log.Printf("Error publishing to %s: %v", provider.Name, err)

				// Save failed post with error information
				post.Status = data_database.PostStatusFailed
				post.ErrorMessage = err.Error()

				if err := db.Create(&post).Error; err != nil {
					log.Printf("Error saving failed post: %v", err)
				}

				results = append(results, fmt.Sprintf("%s: failed to publish", provider.Name))
				hasErrors = true
				continue
			}

			// Extract external ID and URL from postID
			externalID := postID
			externalURL := ""

			// For Facebook, postID is the actual Facebook post ID
			// We can construct the URL: https://www.facebook.com/{postID}
			if provider.Type == "facebook" && postID != "" {
				externalURL = fmt.Sprintf("https://www.facebook.com/%s", postID)
			}

			// Set published time and status
			now := time.Now()
			status := data_database.PostStatusPublished

			// Update post with external information
			post.ExternalID = externalID
			post.ExternalURL = externalURL
			post.PublishedAt = &now
			post.Status = status
			post.ErrorMessage = ""

			// Save post to database
			if err := db.Create(&post).Error; err != nil {
				log.Printf("Error saving post: %v", err)
				results = append(results, fmt.Sprintf("%s: published but not saved", provider.Name))
				hasErrors = true
				continue
			}

			results = append(results, fmt.Sprintf("%s: published successfully (ID: %s)", provider.Name, postID))
		}
	}

	// Prepare response message
	var message string
	if hasErrors {
		message = "Post completed with some errors:\n" + strings.Join(results, "\n")
	} else {
		message = "Post published successfully to all platforms:\n" + strings.Join(results, "\n")
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)

		var responseClass, icon string
		if hasErrors {
			responseClass = "bg-yellow-100 text-yellow-800"
			icon = "⚠️"
		} else {
			responseClass = "bg-green-100 text-green-800"
			icon = "✓"
		}

		responseHTML := fmt.Sprintf(`
			<div class="p-4 %s rounded-lg">
				<div class="font-medium">%s %s</div>
				<pre class="mt-2 text-sm whitespace-pre-wrap">%s</pre>
			</div>
		`, responseClass, icon, "Post completed", message)

		if _, err := w.Write([]byte(responseHTML)); err != nil {
			log.Printf("Error writing response: %v", err)
		}
	} else {
		// Regular form submission - redirect with flash message
		flashType := "success"
		if hasErrors {
			flashType = "warning"
		}
		h.redirectWithFlash(w, r, "/posts", message, flashType)
	}
}

// StatsHandlers for dashboard stats
func (h *WebHandler) HandleProvidersCount(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	db, err := h.dbManager.GetDB(userID)
	if err != nil {
		if _, writeErr := w.Write([]byte("0")); writeErr != nil {
			log.Printf("Error writing response: %v", writeErr)
		}
		return
	}

	var count int64
	db.Model(&data_database.Provider{}).Count(&count)
	if _, err := w.Write([]byte(fmt.Sprintf("%d", count))); err != nil {
		log.Printf("Error writing provider count: %v", err)
	}
}

func (h *WebHandler) HandlePublishedCount(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	db, err := h.dbManager.GetDB(userID)
	if err != nil {
		if _, writeErr := w.Write([]byte("0")); writeErr != nil {
			log.Printf("Error writing response: %v", writeErr)
		}
		return
	}

	var count int64
	db.Model(&data_database.Post{}).Where("user_id = ?", userID).Count(&count)
	if _, err := w.Write([]byte(fmt.Sprintf("%d", count))); err != nil {
		log.Printf("Error writing published count: %v", err)
	}
}

func (h *WebHandler) HandleScheduledCount(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	db, err := h.dbManager.GetDB(userID)
	if err != nil {
		if _, writeErr := w.Write([]byte("0")); writeErr != nil {
			log.Printf("Error writing response: %v", writeErr)
		}
		return
	}

	var count int64
	db.Model(&data_database.ScheduledJob{}).Where("user_id = ? AND status = ?", userID, "pending").Count(&count)
	if _, err := w.Write([]byte(fmt.Sprintf("%d", count))); err != nil {
		log.Printf("Error writing scheduled count: %v", err)
	}
}

func (h *WebHandler) HandleMonthlyCount(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	db, err := h.dbManager.GetDB(userID)
	if err != nil {
		if _, writeErr := w.Write([]byte("0")); writeErr != nil {
			log.Printf("Error writing response: %v", writeErr)
		}
		return
	}

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	var count int64
	db.Model(&data_database.Post{}).Where("user_id = ? AND created_at >= ? AND created_at <= ?",
		userID, startOfMonth, endOfMonth).Count(&count)
	if _, err := w.Write([]byte(fmt.Sprintf("%d", count))); err != nil {
		log.Printf("Error writing monthly count: %v", err)
	}
}

// HandleProvidersOptions returns HTML checkboxes for provider selection
func (h *WebHandler) HandleProvidersOptions(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)

	// Generate cache key based on user ID
	params := map[string]string{
		"user_id": userID,
	}
	cacheKey := GenerateCacheKey("providers_options", params)

	// Check cache first (cache for 2 minutes for provider options)
	if cached, exists := h.cache.Get(cacheKey); exists {
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("X-Cache", "HIT")
		w.Write([]byte(cached))
		return
	}

	db, err := h.dbManager.GetDB(userID)
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("X-Cache", "MISS")
		component.ProvidersOptionsError().Render(r.Context(), w)
		return
	}

	var providers []data_database.Provider
	if err := db.Find(&providers).Error; err != nil {
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("X-Cache", "MISS")
		component.ProvidersOptionsError().Render(r.Context(), w)
		return
	}

	if len(providers) == 0 {
		var buf strings.Builder
		component.ProvidersOptions([]component.ProviderOption{}).Render(r.Context(), &buf)
		html := buf.String()

		// Cache empty result for 1 minute
		h.cache.Set(cacheKey, html, time.Minute)

		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("X-Cache", "MISS")
		w.Write([]byte(html))
		return
	}

	// Convert to component data
	var providerOptions []component.ProviderOption
	for _, provider := range providers {
		// Check if provider is configured
		configured, err := h.providerService.IsProviderConfigured(userID, provider.Name)
		if err != nil {
			configured = false
		}

		// Get provider icon class based on type
		iconClass := "bg-gray-500"
		switch provider.Type {
		case "tiktok":
			iconClass = "bg-black"
		case "instagram":
			iconClass = "bg-gradient-to-r from-purple-500 to-pink-500"
		case "facebook":
			iconClass = "bg-blue-600"
		}

		providerOptions = append(providerOptions, component.ProviderOption{
			ID:         provider.ID,
			Type:       provider.Type,
			Name:       provider.Name,
			IconClass:  iconClass,
			Configured: configured,
		})
	}

	// Render and cache
	var buf strings.Builder
	component.ProvidersOptions(providerOptions).Render(r.Context(), &buf)
	html := buf.String()

	// Cache for 2 minutes
	h.cache.Set(cacheKey, html, 2*time.Minute)

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("X-Cache", "MISS")
	w.Write([]byte(html))
}

// HandleProviderSettings returns HTML for provider-specific settings
func (h *WebHandler) HandleProviderSettings(w http.ResponseWriter, r *http.Request) {
	providerType := r.URL.Query().Get("type")
	providerName := r.URL.Query().Get("name")

	if providerType == "" || providerName == "" {
		http.Error(w, "Provider type and name are required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	switch providerType {
	case "tiktok":
		err := social.TikTokSettings(providerName).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering TikTokSettings: %v", err)
		}
	case "instagram":
		err := social.InstagramSettings(providerName).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering InstagramSettings: %v", err)
		}
	case "facebook":
		err := social.FacebookSettings(providerName).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering FacebookSettings: %v", err)
		}
	default:
		err := social.DefaultProviderSettings(providerName).Render(r.Context(), w)
		if err != nil {
			log.Printf("Error rendering DefaultProviderSettings: %v", err)
		}
	}
}

// HandleProviderTabs returns HTML for provider tabs
func (h *WebHandler) HandleProviderTabs(w http.ResponseWriter, r *http.Request) {
	selectedProviders := r.URL.Query().Get("providers")
	activeIndex := r.URL.Query().Get("active")

	if selectedProviders == "" {
		http.Error(w, "No providers selected", http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)
	db, err := h.dbManager.GetDB(userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Parse selected providers (comma-separated)
	providerIDStrings := strings.Split(selectedProviders, ",")
	var tabs []component.ProviderTab

	for i, providerIDStr := range providerIDStrings {
		providerID, err := strconv.ParseUint(providerIDStr, 10, 32)
		if err != nil {
			continue
		}

		// Fetch provider details from database
		var provider data_database.Provider
		if err := db.First(&provider, providerID).Error; err != nil {
			continue
		}

		tabs = append(tabs, component.ProviderTab{
			Index: i,
			Type:  provider.Type,
			Name:  provider.Name,
		})
	}

	activeIdx := 0
	if activeIndex != "" {
		if idx, err := strconv.Atoi(activeIndex); err == nil && idx >= 0 && idx < len(tabs) {
			activeIdx = idx
		}
	}

	w.Header().Set("Content-Type", "text/html")
	component.ProviderTabs(tabs, activeIdx).Render(r.Context(), w)
}

// HandlePostSuccess returns HTML for post success message
func (h *WebHandler) HandlePostSuccess(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	component.PostSuccess().Render(r.Context(), w)
}

// HandleErrorMessage returns HTML for error message
func (h *WebHandler) HandleErrorMessage(w http.ResponseWriter, r *http.Request) {
	messageType := r.URL.Query().Get("type")
	message := r.URL.Query().Get("message")

	// Generate cache key
	params := map[string]string{
		"type":    messageType,
		"message": message,
	}
	cacheKey := GenerateCacheKey("error_message", params)

	// Check cache first (cache for 5 minutes for error messages)
	if cached, exists := h.cache.Get(cacheKey); exists {
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("X-Cache", "HIT")
		w.Write([]byte(cached))
		return
	}

	// Create message data
	var msgData component.MessageData
	switch messageType {
	case "error":
		msgData.Type = component.MessageTypeError
		msgData.Text = message
		msgData.Default = message == ""
	case "success":
		msgData.Type = component.MessageTypeSuccess
		msgData.Text = message
		msgData.Default = false
	case "info":
		msgData.Type = component.MessageTypeInfo
		msgData.Text = message
		msgData.Default = false
	case "warning":
		msgData.Type = component.MessageTypeWarning
		msgData.Text = message
		msgData.Default = false
	case "loading":
		msgData.Type = component.MessageTypeError
		msgData.Text = "Błąd podczas ładowania"
		msgData.Default = false
	case "success_error":
		msgData.Type = component.MessageTypeError
		msgData.Text = "Błąd podczas ładowania komunikatu sukcesu"
		msgData.Default = false
	default:
		msgData.Type = component.MessageTypeError
		msgData.Text = ""
		msgData.Default = true
	}

	// Render and cache
	var buf strings.Builder
	component.Message(msgData).Render(r.Context(), &buf)
	html := buf.String()

	h.cache.Set(cacheKey, html, 5*time.Minute)

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("X-Cache", "MISS")
	w.Write([]byte(html))
}

// HandleCacheClear clears the component cache
func (h *WebHandler) HandleCacheClear(w http.ResponseWriter, r *http.Request) {
	h.cache.Clear()
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "success", "message": "Cache cleared"}`))
}

// HandleCacheStats returns cache statistics
func (h *WebHandler) HandleCacheStats(w http.ResponseWriter, r *http.Request) {
	// This would need to be implemented in the cache to track stats
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "success", "message": "Cache stats endpoint"}`))
}

// HandleFilePreview returns HTML for file preview
func (h *WebHandler) HandleFilePreview(w http.ResponseWriter, r *http.Request) {
	// This endpoint would receive file data and return preview HTML
	// For now, it's a placeholder that would be called by JavaScript
	w.Header().Set("Content-Type", "text/html")

	// In a real implementation, you'd parse the file data from the request
	// and return the appropriate preview HTML
	component.FilePreviewContainer([]component.FilePreview{}).Render(r.Context(), w)
}

func (h *WebHandler) getUserID(r *http.Request) string {
	// TODO: Implement proper user authentication
	return "default_user"
}

// handleMediaUpload handles the upload of media files
func (h *WebHandler) handleMediaUpload(files []*multipart.FileHeader, userID string) ([]data_database.Media, error) {
	var mediaFiles []data_database.Media

	// Create upload directory
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	for _, fileHeader := range files {
		// Validate file size (10MB max)
		if fileHeader.Size > 10<<20 {
			return nil, fmt.Errorf("file %s is too large (max 10MB)", fileHeader.Filename)
		}

		// Validate file type
		contentType := fileHeader.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "image/") && !strings.HasPrefix(contentType, "video/") {
			return nil, fmt.Errorf("file %s has unsupported type: %s", fileHeader.Filename, contentType)
		}

		// Generate unique filename with UUID
		fileExt := filepath.Ext(fileHeader.Filename)
		uniqueID := uuid.New().String()
		filename := fmt.Sprintf("%s%s", uniqueID, fileExt)
		filePath := filepath.Join(uploadDir, filename)

		// Save file
		file, err := fileHeader.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open uploaded file: %w", err)
		}
		defer file.Close()

		dst, err := os.Create(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create destination file: %w", err)
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			return nil, fmt.Errorf("failed to save file: %w", err)
		}

		// Determine file type
		fileType := "image"
		if strings.HasPrefix(contentType, "video/") {
			fileType = "video"
		}

		// Create media record
		media := data_database.Media{
			FileName:  filename, // Store the UUID filename
			FilePath:  filePath,
			FileType:  fileType,
			FileSize:  fileHeader.Size,
			MimeType:  contentType,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mediaFiles = append(mediaFiles, media)
	}

	return mediaFiles, nil
}

// extractProviderSettings extracts provider-specific settings from form data
func (h *WebHandler) extractProviderSettings(form url.Values) map[string]map[string]string {
	settings := make(map[string]map[string]string)

	// Extract TikTok settings
	tiktokSettings := make(map[string]string)
	for key, values := range form {
		if strings.HasPrefix(key, "tiktok_") && len(values) > 0 {
			// Extract provider name from key (e.g., "tiktok_visibility_providername")
			parts := strings.Split(key, "_")
			if len(parts) >= 3 {
				settingType := parts[1]
				providerName := strings.Join(parts[2:], "_")
				tiktokSettings[settingType] = values[0]
				if settings[providerName] == nil {
					settings[providerName] = make(map[string]string)
				}
				settings[providerName]["tiktok_"+settingType] = values[0]
			}
		}
	}

	// Extract Instagram settings
	for key, values := range form {
		if strings.HasPrefix(key, "instagram_") && len(values) > 0 {
			parts := strings.Split(key, "_")
			if len(parts) >= 3 {
				settingType := parts[1]
				providerName := strings.Join(parts[2:], "_")
				if settings[providerName] == nil {
					settings[providerName] = make(map[string]string)
				}
				settings[providerName]["instagram_"+settingType] = values[0]
			}
		}
	}

	// Extract Facebook settings
	for key, values := range form {
		if strings.HasPrefix(key, "facebook_") && len(values) > 0 {
			parts := strings.Split(key, "_")
			if len(parts) >= 3 {
				settingType := parts[1]
				providerName := strings.Join(parts[2:], "_")
				if settings[providerName] == nil {
					settings[providerName] = make(map[string]string)
				}
				settings[providerName]["facebook_"+settingType] = values[0]
			}
		}
	}

	return settings
}

// SettingsPage handles the settings page
func (h *WebHandler) SettingsPage(w http.ResponseWriter, r *http.Request) {
	// Get flash message from query parameters
	flashMessage := ""
	flashType := "info"
	if flashMsg := r.URL.Query().Get("flash"); flashMsg != "" {
		if decoded, err := url.QueryUnescape(flashMsg); err == nil {
			flashMessage = decoded
		} else {
			flashMessage = flashMsg
		}
		flashType = r.URL.Query().Get("flash_type")
		if flashType == "" {
			flashType = "info"
		}
	}

	// Get active section from query parameters (default to "user")
	activeSection := r.URL.Query().Get("section")
	if activeSection == "" {
		activeSection = "user"
	}

	// Get user tokens for user section
	var userTokens []component.UserToken
	if activeSection == "user" {
		userID := h.getUserID(r)
		db, err := h.dbManager.GetDB(userID)
		if err == nil {
			var apiTokens []data_database.APIToken
			if err := db.Where("user_id = ?", userID).Find(&apiTokens).Error; err == nil {
				for _, token := range apiTokens {
					userToken := component.UserToken{
						ID:        fmt.Sprintf("%d", token.ID),
						Name:      fmt.Sprintf("Token #%d", token.ID),
						CreatedAt: token.CreatedAt.Format("2006-01-02 15:04:05"),
						LastUsed:  "",
					}
					if token.LastUsed != nil {
						userToken.LastUsed = token.LastUsed.Format("2006-01-02 15:04:05")
					}
					userTokens = append(userTokens, userToken)
				}
			}
		}
	}

	// Create settings data
	settingsData := templates.SettingsData{
		ActiveSection: activeSection,
		UserTokens:    userTokens,
	}

	// Create layout data
	layoutData := templates.LayoutData{
		Title:        "Ustawienia",
		CurrentPage:  "settings",
		FlashMessage: flashMessage,
		FlashType:    flashType,
		Content:      templates.SettingsContent(settingsData),
	}

	// Render the layout
	w.Header().Set("Content-Type", "text/html")
	layoutComponent := templates.Layout(layoutData)
	if err := layoutComponent.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering settings page: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
