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

	data_database "github.com/tkowalski/socgo/internal/data/database"
	"github.com/tkowalski/socgo/internal/service/database"
	"github.com/tkowalski/socgo/internal/service/post"
	"github.com/tkowalski/socgo/web/templates"
	"github.com/tkowalski/socgo/web/templates/social"
)

// WebHandler handles web page requests
type WebHandler struct {
	dbManager       *database.Manager
	providerService *post.ProviderService
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

// HomePage handles the home page
func (h *WebHandler) HomePage(w http.ResponseWriter, r *http.Request) {
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
		Title:        "Home",
		CurrentPage:  "home",
		FlashMessage: flashMessage,
		FlashType:    flashType,
		Content:      templates.HomeContent(),
	}

	// Render the layout
	w.Header().Set("Content-Type", "text/html")
	layoutComponent := templates.Layout(layoutData)
	if err := layoutComponent.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering home page: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// DashboardPage handles the dashboard page
func (h *WebHandler) DashboardPage(w http.ResponseWriter, r *http.Request) {
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
		Title:        "Dashboard",
		CurrentPage:  "dashboard",
		FlashMessage: flashMessage,
		FlashType:    flashType,
		Content:      templates.DashboardContent(),
	}

	// Render the layout
	w.Header().Set("Content-Type", "text/html")
	layoutComponent := templates.Layout(layoutData)
	if err := layoutComponent.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering dashboard page: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// ProvidersPage handles the providers page
func (h *WebHandler) ProvidersPage(w http.ResponseWriter, r *http.Request) {
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
		Title:        "Providers",
		CurrentPage:  "providers",
		FlashMessage: flashMessage,
		FlashType:    flashType,
		Content:      templates.ProvidersContent(),
	}

	// Render the layout
	w.Header().Set("Content-Type", "text/html")
	layoutComponent := templates.Layout(layoutData)
	if err := layoutComponent.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering providers page: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// PostsPage handles the posts page
func (h *WebHandler) PostsPage(w http.ResponseWriter, r *http.Request) {
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
		Title:        "Create Post",
		CurrentPage:  "posts",
		FlashMessage: flashMessage,
		FlashType:    flashType,
		Content:      templates.PostsContent(),
	}

	// Render the layout
	w.Header().Set("Content-Type", "text/html")
	layoutComponent := templates.Layout(layoutData)
	if err := layoutComponent.Render(r.Context(), w); err != nil {
		log.Printf("Error rendering posts page: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
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
		if settings != nil {
			log.Printf("Provider %s settings: %v", provider.Name, settings)
		}

		// Create post record
		post := data_database.Post{
			Content:    content,
			UserID:     userID,
			ProviderID: uint(providerID),
			Media:      mediaFiles,
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
			postID, err := h.providerService.PublishContent(ctx, userID, provider.Name, content)
			if err != nil {
				log.Printf("Error publishing to %s: %v", provider.Name, err)
				results = append(results, fmt.Sprintf("%s: failed to publish", provider.Name))
				hasErrors = true
				continue
			}

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

// responseCapture captures response data
type responseCapture struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (rc *responseCapture) WriteHeader(statusCode int) {
	rc.statusCode = statusCode
}

func (rc *responseCapture) Write(data []byte) (int, error) {
	rc.body = append(rc.body, data...)
	return len(data), nil
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
	db, err := h.dbManager.GetDB(userID)
	if err != nil {
		if _, writeErr := w.Write([]byte(`<p class="text-red-600">Error loading providers</p>`)); writeErr != nil {
			log.Printf("Error writing response: %v", writeErr)
		}
		return
	}

	var providers []data_database.Provider
	if err := db.Find(&providers).Error; err != nil {
		if _, writeErr := w.Write([]byte(`<p class="text-red-600">Error loading providers</p>`)); writeErr != nil {
			log.Printf("Error writing response: %v", writeErr)
		}
		return
	}

	if len(providers) == 0 {
		w.Header().Set("Content-Type", "text/html")
		if _, err := w.Write([]byte(`<p class="text-gray-500">No providers available. <a href="/providers" class="text-blue-600 hover:underline">Connect your accounts first</a>.</p>`)); err != nil {
			log.Printf("Error writing providers options: %v", err)
		}
		return
	}

	html := `<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">`
	for _, provider := range providers {
		// Check if provider is configured
		configured, err := h.providerService.IsProviderConfigured(userID, provider.Name)
		if err != nil || !configured {
			html += fmt.Sprintf(`
				<label class="flex items-center p-3 border border-gray-200 rounded-lg cursor-not-allowed opacity-50">
					<input type="checkbox" name="providers" value="%d" disabled class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded">
					<div class="ml-3">
						<div class="text-sm font-medium text-gray-700">%s</div>
						<div class="text-xs text-gray-500">Not connected</div>
					</div>
				</label>
			`, provider.ID, provider.Name)
		} else {
			html += fmt.Sprintf(`
				<label class="flex items-center p-3 border border-gray-200 rounded-lg hover:bg-gray-50 cursor-pointer">
					<input type="checkbox" name="providers" value="%d" data-provider-type="%s" data-provider-name="%s" class="provider-checkbox h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded">
					<div class="ml-3">
						<div class="text-sm font-medium text-gray-700">%s</div>
						<div class="text-xs text-green-600">Connected</div>
					</div>
				</label>
			`, provider.ID, provider.Type, provider.Name, provider.Name)
		}
	}
	html += `</div>`

	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(html)); err != nil {
		log.Printf("Error writing providers options: %v", err)
	}
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

func (h *WebHandler) getUserID(r *http.Request) string {
	// TODO: Implement proper user authentication
	return "default_user"
}

// handleMediaUpload handles the upload of media files
func (h *WebHandler) handleMediaUpload(files []*multipart.FileHeader, userID string) ([]data_database.Media, error) {
	var mediaFiles []data_database.Media

	// Create upload directory for user
	uploadDir := filepath.Join("uploads", userID)
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

		// Generate unique filename
		filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(fileHeader.Filename))
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
			FileName:  fileHeader.Filename,
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
