package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	data_database "github.com/tkowalski/socgo/internal/data/database"
	provider_pkg "github.com/tkowalski/socgo/internal/data/provider"
	"github.com/tkowalski/socgo/internal/handlers/internal"
	process_post "github.com/tkowalski/socgo/internal/process/post"
	process_post_data "github.com/tkowalski/socgo/internal/process/post/data"
	"github.com/tkowalski/socgo/internal/service/database"
	"github.com/tkowalski/socgo/internal/service/post"
	"github.com/tkowalski/socgo/web/templates/component"
	"github.com/tkowalski/socgo/web/templates/social"
)

type PostRequest struct {
	ProviderID uint   `json:"provider_id"`
	Content    string `json:"content"`
	ScheduleAt string `json:"schedule_at"` // ISO8601 format or "now"
}

type PostResponse struct {
	ID         uint      `json:"id"`
	Status     string    `json:"status"`
	ProviderID uint      `json:"provider_id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	Message    string    `json:"message,omitempty"`
}

type HistoryPost struct {
	ID          uint                   `json:"id"`
	Content     string                 `json:"content"`
	ProviderID  uint                   `json:"provider_id"`
	Provider    data_database.Provider `json:"provider"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	Status      string                 `json:"status"`
}

type HistoryResponse struct {
	Posts []HistoryPost `json:"posts"`
	Page  int           `json:"page"`
	Total int64         `json:"total"`
}

type PostHandler struct {
	dbManager       *database.Manager
	providerService *post.ProviderService
}

func NewPostHandler(dbManager *database.Manager, providerService *post.ProviderService) *PostHandler {
	return &PostHandler{
		dbManager:       dbManager,
		providerService: providerService,
	}
}

// HandleFilePreview returns HTML for file preview
func (h *PostHandler) HandleFilePreview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	// Parse JSON data from request body
	var requestData struct {
		Files []struct {
			Index    int    `json:"index"`
			FileName string `json:"fileName"`
			FileType string `json:"fileType"`
			DataURL  string `json:"dataURL"`
		} `json:"files"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	log.Printf("Received %d files for preview", len(requestData.Files))

	// Convert to component data
	var filePreviews []component.FilePreview
	for _, file := range requestData.Files {
		filePreviews = append(filePreviews, component.FilePreview{
			Index:    file.Index,
			FileName: file.FileName,
			FileType: file.FileType,
			DataURL:  file.DataURL,
		})
		log.Printf("File: %s, Type: %s", file.FileName, file.FileType)
	}

	// Render the preview HTML
	if err := component.FilePreviewContainer(filePreviews).Render(r.Context(), w); err != nil {
		log.Printf("Error rendering file preview: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *PostHandler) HandlePost(w http.ResponseWriter, r *http.Request) {
	// Create post context
	ctx := &process_post_data.PostContext{
		Request:     r,
		UserID:      internal.GetUserID(r),
		DB:          h.dbManager,
		Content:     "",
		ScheduleAt:  "",
		Media:       []provider_pkg.Media{},
		Settings:    make(map[string]string),
		ProviderIDs: []uint{},
		PayloadJSON: []byte{},
		Errors:      []error{},
	}

	postProcess := process_post.NewPostProcess(h.providerService)

	// Execute the process
	if err := postProcess.Execute(ctx); err != nil {
		log.Printf("Post process failed: %v", err)

		// Return HTML error for HTMX
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadRequest)
		errorHTML := fmt.Sprintf(`<div class="text-red-600 text-center py-8">
			<h3 class="text-lg font-semibold mb-2">Błąd przetwarzania</h3>
			<p>%s</p>
		</div>`, err.Error())
		if _, writeErr := w.Write([]byte(errorHTML)); writeErr != nil {
			log.Printf("Error writing error response: %v", writeErr)
		}
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "text/html")
	if err := component.PostSuccess().Render(r.Context(), w); err != nil {
		log.Printf("Error rendering post success: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *PostHandler) HandleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get page parameter (default to 1)
	pageStr := r.URL.Query().Get("page")
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	userID := internal.GetUserID(r)
	db, err := h.dbManager.GetDB(userID)
	if err != nil {
		log.Printf("Error getting database: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	const pageSize = 20
	offset := (page - 1) * pageSize

	// Get total count
	var total int64
	if err := db.Model(&data_database.Post{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		log.Printf("Error counting posts: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get posts with pagination and include scheduled jobs
	var posts []data_database.Post
	if err := db.Preload("Provider").Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(pageSize).Offset(offset).
		Find(&posts).Error; err != nil {
		log.Printf("Error fetching posts: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get scheduled jobs
	var scheduledJobs []data_database.ScheduledJob
	if err := db.Preload("Provider").Where("user_id = ?", userID).
		Order("scheduled_at DESC").
		Limit(pageSize).Offset(offset).
		Find(&scheduledJobs).Error; err != nil {
		log.Printf("Error fetching scheduled jobs: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert to history posts
	var historyPosts []HistoryPost

	// Add published posts
	for _, post := range posts {
		historyPosts = append(historyPosts, HistoryPost{
			ID:         post.ID,
			Content:    post.Content,
			ProviderID: post.ProviderID,
			Provider:   post.Provider,
			CreatedAt:  post.CreatedAt,
			Status:     "published",
		})
	}

	// Add scheduled posts
	for _, job := range scheduledJobs {
		// Parse PayloadData to extract content
		var payloadData struct {
			Content string `json:"content"`
		}

		// Try to parse as JSON first, if it fails, treat as plain text
		if err := json.Unmarshal([]byte(job.PayloadData), &payloadData); err != nil {
			// If it's not JSON, treat as plain text content
			payloadData.Content = job.PayloadData
		}

		historyPosts = append(historyPosts, HistoryPost{
			ID:          job.ID,
			Content:     payloadData.Content,
			ProviderID:  job.ProviderID,
			Provider:    job.Provider,
			ScheduledAt: &job.ScheduledAt,
			CreatedAt:   job.CreatedAt,
			Status:      job.Status,
		})
	}

	// Check if request wants JSON (API) or HTML (HTMX)
	if r.Header.Get("Accept") == "application/json" {
		response := HistoryResponse{
			Posts: historyPosts,
			Page:  page,
			Total: total,
		}
		h.writeJSONResponse(w, response, http.StatusOK)
		return
	}

	// Return HTML for HTMX
	var htmlBuilder strings.Builder
	htmlBuilder.WriteString(`<div class="space-y-4">`)

	for _, post := range historyPosts {
		htmlBuilder.WriteString(`<div class="bg-white rounded-lg shadow p-4 border-l-4 border-blue-500">`)
		htmlBuilder.WriteString(`<div class="flex justify-between items-start">`)
		htmlBuilder.WriteString(`<div class="flex-1">`)
		htmlBuilder.WriteString(`<div class="flex items-center space-x-2 mb-2">`)
		htmlBuilder.WriteString(`<span class="text-sm font-medium text-gray-500">`)
		htmlBuilder.WriteString(post.Provider.Name)
		htmlBuilder.WriteString(`</span>`)
		htmlBuilder.WriteString(`<span class="px-2 py-1 text-xs rounded-full `)
		if post.Status == "published" {
			htmlBuilder.WriteString(`bg-green-100 text-green-800`)
		} else {
			htmlBuilder.WriteString(`bg-yellow-100 text-yellow-800`)
		}
		htmlBuilder.WriteString(`">`)
		htmlBuilder.WriteString(post.Status)
		htmlBuilder.WriteString(`</span>`)
		htmlBuilder.WriteString(`</div>`)
		htmlBuilder.WriteString(`<p class="text-gray-800 mb-2">`)
		htmlBuilder.WriteString(h.truncateText(post.Content, 200))
		htmlBuilder.WriteString(`</p>`)
		htmlBuilder.WriteString(`<div class="text-sm text-gray-500">`)
		if post.ScheduledAt != nil {
			htmlBuilder.WriteString(`Scheduled for: `)
			htmlBuilder.WriteString(post.ScheduledAt.Format("Jan 2, 2006 at 3:04 PM"))
		} else {
			htmlBuilder.WriteString(`Published: `)
			htmlBuilder.WriteString(post.CreatedAt.Format("Jan 2, 2006 at 3:04 PM"))
		}
		htmlBuilder.WriteString(`</div>`)
		htmlBuilder.WriteString(`</div>`)
		htmlBuilder.WriteString(`<div class="flex space-x-2">`)
		htmlBuilder.WriteString(`<button onclick="viewPostDetails(`)
		htmlBuilder.WriteString(strconv.FormatUint(uint64(post.ID), 10))
		htmlBuilder.WriteString(`)" class="text-blue-600 hover:text-blue-800 text-sm">View</button>`)
		if post.Status != "published" {
			htmlBuilder.WriteString(`<button onclick="deletePost(`)
			htmlBuilder.WriteString(strconv.FormatUint(uint64(post.ID), 10))
			htmlBuilder.WriteString(`)" class="text-red-600 hover:text-red-800 text-sm">Delete</button>`)
		}
		htmlBuilder.WriteString(`</div>`)
		htmlBuilder.WriteString(`</div>`)
		htmlBuilder.WriteString(`</div>`)
	}

	htmlBuilder.WriteString(`</div>`)

	// Add pagination if there are more posts
	if int64(page*pageSize) < total {
		htmlBuilder.WriteString(`<div class="mt-4 text-center">`)
		htmlBuilder.WriteString(`<button hx-get="/posts/history?page=`)
		htmlBuilder.WriteString(strconv.Itoa(page + 1))
		htmlBuilder.WriteString(`" hx-target="#history-list" class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded">`)
		htmlBuilder.WriteString(`Load More`)
		htmlBuilder.WriteString(`</button>`)
		htmlBuilder.WriteString(`</div>`)
	}

	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(htmlBuilder.String())); err != nil {
		log.Printf("Error writing history response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *PostHandler) truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength-3] + "..."
}

func (h *PostHandler) HandlePostDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract post ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	postIDStr := pathParts[3]
	postType := r.URL.Query().Get("type")

	postID, err := strconv.ParseUint(postIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	userID := internal.GetUserID(r)
	db, err := h.dbManager.GetDB(userID)
	if err != nil {
		log.Printf("Error getting database: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var htmlBuilder strings.Builder

	if postType == "scheduled" {
		// Get scheduled job
		var job data_database.ScheduledJob
		if err := db.Preload("Provider").First(&job, postID).Error; err != nil {
			http.Error(w, "Scheduled post not found", http.StatusNotFound)
			return
		}

		statusClass := "bg-green-100 text-green-800"
		if job.Status == "pending" {
			statusClass = "bg-yellow-100 text-yellow-800"
		} else if job.Status == "failed" {
			statusClass = "bg-red-100 text-red-800"
		}

		// Parse PayloadData to check if it contains media information
		var payloadData struct {
			Content string `json:"content"`
			Media   []struct {
				FileName string `json:"file_name"`
				FileType string `json:"file_type"`
				FilePath string `json:"file_path"`
			} `json:"media"`
		}

		// Try to parse as JSON first, if it fails, treat as plain text
		if err := json.Unmarshal([]byte(job.PayloadData), &payloadData); err != nil {
			// If it's not JSON, treat as plain text content
			payloadData.Content = job.PayloadData
		}

		htmlBuilder.WriteString(fmt.Sprintf(`
			<div class="space-y-4">
				<div class="flex justify-between items-start">
					<div>
						<h4 class="text-lg font-semibold text-gray-900">Zaplanowany post</h4>
						<p class="text-sm text-gray-600">ID: %d</p>
					</div>
					<span class="px-3 py-1 text-sm rounded-full %s">%s</span>
				</div>
				
				<div class="bg-gray-50 p-4 rounded-lg">
					<h5 class="font-medium text-gray-900 mb-2">Treść:</h5>
					<p class="text-gray-800 whitespace-pre-wrap">%s</p>
				</div>
		`, job.ID, statusClass, job.Status, payloadData.Content))

		// Add media section if media exists
		if len(payloadData.Media) > 0 {
			htmlBuilder.WriteString(`
				<div class="bg-gray-50 p-4 rounded-lg">
					<h5 class="font-medium text-gray-900 mb-2">Media:</h5>
					<div class="grid grid-cols-2 gap-2">
			`)

			for _, media := range payloadData.Media {
				if strings.Contains(media.FileType, "image") {
					htmlBuilder.WriteString(fmt.Sprintf(`
						<div class="relative">
							<img src="/uploads/%s" alt="%s" class="w-full h-24 object-cover rounded-lg">
							<p class="text-xs text-gray-600 mt-1 truncate">%s</p>
						</div>
					`, filepath.Base(media.FilePath), media.FileName, media.FileName))
				} else if strings.Contains(media.FileType, "video") {
					htmlBuilder.WriteString(fmt.Sprintf(`
						<div class="relative">
							<video src="/uploads/%s" class="w-full h-24 object-cover rounded-lg" controls muted>
								Twoja przeglądarka nie wspiera odtwarzania wideo.
							</video>
							<p class="text-xs text-gray-600 mt-1 truncate">%s</p>
						</div>
					`, filepath.Base(media.FilePath), media.FileName))
				}
			}

			htmlBuilder.WriteString(`
					</div>
				</div>
			`)
		}

		htmlBuilder.WriteString(fmt.Sprintf(`
				<div class="grid grid-cols-2 gap-4">
					<div>
						<h5 class="font-medium text-gray-900 mb-1">Platforma:</h5>
						<p class="text-gray-600">%s</p>
					</div>
					<div>
						<h5 class="font-medium text-gray-900 mb-1">Zaplanowany na:</h5>
						<p class="text-gray-600">%s</p>
					</div>
					<div>
						<h5 class="font-medium text-gray-900 mb-1">Utworzony:</h5>
						<p class="text-gray-600">%s</p>
					</div>
					<div>
						<h5 class="font-medium text-gray-900 mb-1">Status:</h5>
						<p class="text-gray-600">%s</p>
					</div>
				</div>
				
				<div class="flex space-x-2">
					<button onclick="closePostDetails()" class="bg-gray-500 hover:bg-gray-700 text-white font-bold py-2 px-4 rounded">
						Zamknij
					</button>
					<button onclick="deletePost(%d)" class="bg-red-500 hover:bg-red-700 text-white font-bold py-2 px-4 rounded">
						Usuń
					</button>
				</div>
			</div>
		`, job.Provider.Name, h.formatPublishedAt(&job.ScheduledAt), h.formatPublishedAt(&job.CreatedAt), job.Status, job.ID))
	} else {
		// Get published post with media
		var post data_database.Post
		if err := db.Preload("Provider").Preload("Media").First(&post, postID).Error; err != nil {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}

		htmlBuilder.WriteString(fmt.Sprintf(`
			<div class="space-y-4">
				<div class="flex justify-between items-start">
					<div>
						<h4 class="text-lg font-semibold text-gray-900">Opublikowany post</h4>
						<p class="text-sm text-gray-600">ID: %d</p>
						<p class="text-sm text-gray-700 font-medium mt-1">Platforma: <span class="text-blue-700">%s</span></p>
						%s
					</div>
					<span class="px-3 py-1 text-sm rounded-full bg-green-100 text-green-800">Opublikowany</span>
				</div>
				
				<div class="bg-gray-50 p-4 rounded-lg">
					<h5 class="font-medium text-gray-900 mb-2">Treść:</h5>
					<p class="text-gray-800 whitespace-pre-wrap">%s</p>
				</div>
		`, post.ID, post.Provider.Name, func() string {
			if post.ExternalID != "" || post.ExternalURL != "" {
				idRow := ""
				if post.ExternalID != "" {
					idRow += fmt.Sprintf(`<p class=\"text-sm text-gray-700\">ID na platformie: <span class=\"font-mono\">%s</span></p>`, post.ExternalID)
				}
				if post.ExternalURL != "" {
					idRow += fmt.Sprintf(`<p class=\"text-sm text-gray-700\">Link: <a href=\"%s\" target=\"_blank\" class=\"text-blue-600 hover:text-blue-800 underline break-all\">%s</a></p>`, post.ExternalURL, post.ExternalURL)
				}
				return idRow
			}
			return ""
		}(), post.Content))

		// Add media section if media exists
		if len(post.Media) > 0 {
			htmlBuilder.WriteString(`
				<div class="bg-gray-50 p-4 rounded-lg">
					<h5 class="font-medium text-gray-900 mb-2">Media:</h5>
					<div class="grid grid-cols-2 gap-2">
			`)

			for _, media := range post.Media {
				if strings.Contains(media.FileType, "image") {
					htmlBuilder.WriteString(fmt.Sprintf(`
						<div class="relative">
							<img src="/uploads/%s" alt="%s" class="w-full h-24 object-cover rounded-lg">
							<p class="text-xs text-gray-600 mt-1 truncate">%s</p>
						</div>
					`, filepath.Base(media.FilePath), media.FileName, media.FileName))
				} else if strings.Contains(media.FileType, "video") {
					htmlBuilder.WriteString(fmt.Sprintf(`
						<div class="relative">
							<video src="/uploads/%s" class="w-full h-24 object-cover rounded-lg" controls muted>
								Twoja przeglądarka nie wspiera odtwarzania wideo.
							</video>
							<p class="text-xs text-gray-600 mt-1 truncate">%s</p>
						</div>
					`, filepath.Base(media.FilePath), media.FileName))
				}
			}

			htmlBuilder.WriteString(`
					</div>
				</div>
			`)
		}

		// Add external link if available
		if post.ExternalURL != "" {
			htmlBuilder.WriteString(fmt.Sprintf(`
				<div class="bg-blue-50 p-4 rounded-lg">
					<h5 class="font-medium text-gray-900 mb-2">Link do posta:</h5>
					<a href="%s" target="_blank" class="text-blue-600 hover:text-blue-800 underline break-all">%s</a>
				</div>
			`, post.ExternalURL, post.ExternalURL))
		}

		htmlBuilder.WriteString(fmt.Sprintf(`
				<div class="grid grid-cols-2 gap-4">
					<div>
						<h5 class="font-medium text-gray-900 mb-1">Opublikowany:</h5>
						<p class="text-gray-600">%s</p>
					</div>
					<div>
						<h5 class="font-medium text-gray-900 mb-1">Utworzony:</h5>
						<p class="text-gray-600">%s</p>
					</div>
					<div>
						<h5 class="font-medium text-gray-900 mb-1">Status:</h5>
						<p class="text-gray-600">Opublikowany</p>
					</div>
				</div>
			</div>
		`, h.formatPublishedAt(&post.CreatedAt), h.formatPublishedAt(&post.CreatedAt)))
	}

	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(htmlBuilder.String())); err != nil {
		log.Printf("Error writing post details response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *PostHandler) formatPublishedAt(publishedAt *time.Time) string {
	if publishedAt == nil {
		return "N/A"
	}
	return publishedAt.Format("Jan 2, 2006 at 3:04 PM")
}

func (h *PostHandler) HandleDeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract post ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	postIDStr := pathParts[3]
	postType := r.URL.Query().Get("type")

	postID, err := strconv.ParseUint(postIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	userID := internal.GetUserID(r)
	db, err := h.dbManager.GetDB(userID)
	if err != nil {
		log.Printf("Error getting database: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if postType == "scheduled" {
		// Delete scheduled job
		var job data_database.ScheduledJob
		if err := db.First(&job, postID).Error; err != nil {
			http.Error(w, "Scheduled post not found", http.StatusNotFound)
			return
		}

		if err := db.Delete(&job).Error; err != nil {
			log.Printf("Error deleting scheduled job: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		// Delete published post
		var post data_database.Post
		if err := db.First(&post, postID).Error; err != nil {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}

		if err := db.Delete(&post).Error; err != nil {
			log.Printf("Error deleting post: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	// Return success response
	response := map[string]string{
		"status":  "success",
		"message": "Post deleted successfully",
	}
	h.writeJSONResponse(w, response, http.StatusOK)
}

func (h *PostHandler) writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// HandlePostSuccess returns HTML for post success message
func (h *PostHandler) HandlePostSuccess(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if err := component.PostSuccess().Render(r.Context(), w); err != nil {
		log.Printf("Error rendering post success: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// HandleProvidersCount handles the providers count
func (h *PostHandler) HandleProvidersCount(w http.ResponseWriter, r *http.Request) {
	userID := internal.GetUserID(r)
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

func (h *PostHandler) HandlePublishedCount(w http.ResponseWriter, r *http.Request) {
	userID := internal.GetUserID(r)
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

func (h *PostHandler) HandleScheduledCount(w http.ResponseWriter, r *http.Request) {
	userID := internal.GetUserID(r)
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

func (h *PostHandler) HandleMonthlyCount(w http.ResponseWriter, r *http.Request) {
	userID := internal.GetUserID(r)
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
func (h *PostHandler) HandleProvidersOptions(w http.ResponseWriter, r *http.Request) {
	userID := internal.GetUserID(r)

	db, err := h.dbManager.GetDB(userID)
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		if err := component.ProvidersOptionsError().Render(r.Context(), w); err != nil {
			log.Printf("Error rendering providers options error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		return
	}

	var providers []data_database.Provider
	if err := db.Find(&providers).Error; err != nil {
		w.Header().Set("Content-Type", "text/html")
		if err := component.ProvidersOptionsError().Render(r.Context(), w); err != nil {
			log.Printf("Error rendering providers options error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		return
	}

	if len(providers) == 0 {
		var buf strings.Builder
		if err := component.ProvidersOptions([]component.ProviderOption{}).Render(r.Context(), &buf); err != nil {
			log.Printf("Error rendering providers options: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		html := buf.String()

		w.Header().Set("Content-Type", "text/html")
		if _, err := w.Write([]byte(html)); err != nil {
			log.Printf("Error writing providers options: %v", err)
		}
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

	// Render
	var buf strings.Builder
	if err := component.ProvidersOptions(providerOptions).Render(r.Context(), &buf); err != nil {
		log.Printf("Error rendering providers options: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	html := buf.String()

	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(html)); err != nil {
		log.Printf("Error writing providers options: %v", err)
	}
}

// HandleProviderSettings returns HTML for provider-specific settings
func (h *PostHandler) HandleProviderSettings(w http.ResponseWriter, r *http.Request) {
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
func (h *PostHandler) HandleProviderTabs(w http.ResponseWriter, r *http.Request) {
	selectedProviders := r.URL.Query().Get("providers")
	activeIndex := r.URL.Query().Get("active")

	if selectedProviders == "" {
		http.Error(w, "No providers selected", http.StatusBadRequest)
		return
	}

	userID := internal.GetUserID(r)
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
			ID:    provider.ID,
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
	if err := component.ProviderTabs(tabs, activeIdx).Render(r.Context(), w); err != nil {
		log.Printf("Error rendering provider tabs: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
