package handlers

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/tkowalski/socgo/internal/database"
	"github.com/tkowalski/socgo/internal/providers"
)

// Request/Response structs for POST /posts endpoint
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

// PostHandler handles POST requests for creating posts
type PostHandler struct {
	dbManager       *database.Manager
	providerService *providers.ProviderService
}

// NewPostHandler creates a new PostHandler instance
func NewPostHandler(dbManager *database.Manager, providerService *providers.ProviderService) *PostHandler {
	return &PostHandler{
		dbManager:       dbManager,
		providerService: providerService,
	}
}

func (h *PostHandler) HandlePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.ProviderID == 0 {
		http.Error(w, "provider_id is required", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Content) == "" {
		http.Error(w, "content is required", http.StatusBadRequest)
		return
	}
	if req.ScheduleAt == "" {
		req.ScheduleAt = "now"
	}

	// Get user ID (currently defaults to "default_user")
	userID := h.getUserID(r)

	// Get database instance for user
	db, err := h.dbManager.GetDB(userID)
	if err != nil {
		log.Printf("Error getting database: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Validate provider exists and is configured
	var provider database.Provider
	if err := db.First(&provider, req.ProviderID).Error; err != nil {
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}

	// Check if provider is configured
	isConfigured, err := h.providerService.IsProviderConfigured(userID, provider.Name)
	if err != nil {
		log.Printf("Error checking provider configuration: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !isConfigured {
		http.Error(w, "Provider not configured", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Handle immediate or scheduled posting
	if req.ScheduleAt == "now" {
		// Immediate posting
		postID, err := h.providerService.PublishContent(ctx, userID, provider.Name, req.Content)
		if err != nil {
			log.Printf("Error publishing content: %v", err)
			http.Error(w, "Failed to publish content", http.StatusInternalServerError)
			return
		}

		// Save post to database
		post := database.Post{
			Content:    req.Content,
			UserID:     userID,
			ProviderID: req.ProviderID,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		if err := db.Create(&post).Error; err != nil {
			log.Printf("Error saving post: %v", err)
			http.Error(w, "Failed to save post", http.StatusInternalServerError)
			return
		}

		// Return success response
		response := PostResponse{
			ID:         post.ID,
			Status:     "published",
			ProviderID: req.ProviderID,
			Content:    req.Content,
			CreatedAt:  post.CreatedAt,
			Message:    "Post published successfully. Post ID: " + postID,
		}

		h.writeJSONResponse(w, response, http.StatusCreated)
	} else {
		// Scheduled posting
		scheduledAt, err := time.Parse(time.RFC3339, req.ScheduleAt)
		if err != nil {
			http.Error(w, "Invalid schedule_at format. Use ISO8601 format or 'now'", http.StatusBadRequest)
			return
		}

		if scheduledAt.Before(time.Now()) {
			http.Error(w, "scheduled_at must be in the future", http.StatusBadRequest)
			return
		}

		// Create scheduled job
		job := database.ScheduledJob{
			JobType:     "publish_post",
			PayloadData: req.Content,
			UserID:      userID,
			ProviderID:  req.ProviderID,
			ScheduledAt: scheduledAt,
			Status:      "pending",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := db.Create(&job).Error; err != nil {
			log.Printf("Error creating scheduled job: %v", err)
			http.Error(w, "Failed to schedule post", http.StatusInternalServerError)
			return
		}

		// Return success response
		response := PostResponse{
			ID:         job.ID,
			Status:     "pending",
			ProviderID: req.ProviderID,
			Content:    req.Content,
			CreatedAt:  job.CreatedAt,
			Message:    "Post scheduled successfully for " + scheduledAt.Format(time.RFC3339),
		}

		h.writeJSONResponse(w, response, http.StatusCreated)
	}
}

func (h *PostHandler) getUserID(r *http.Request) string {
	// TODO: Implement proper user authentication
	return "default_user"
}

func (h *PostHandler) writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

var homeTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SocGo</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <h1 class="text-4xl font-bold text-center text-gray-800 mb-8">
            Welcome to SocGo
        </h1>
        <div class="max-w-md mx-auto bg-white rounded-lg shadow-md p-6">
            <p class="text-gray-600 text-center">
                A simple social media app built with Go, HTMX, and Tailwind CSS.
            </p>
            <button 
                hx-get="/health" 
                hx-target="#status"
                class="mt-4 w-full bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"
            >
                Check Status
            </button>
            <div id="status" class="mt-4 text-center"></div>
        </div>
    </div>
</body>
</html>
`

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("home").Parse(homeTemplate)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, nil); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(`<div class="p-2 bg-green-100 text-green-800 rounded">✓ Server is healthy</div>`)); err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
