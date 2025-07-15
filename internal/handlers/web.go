package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/tkowalski/socgo/internal/database"
	"github.com/tkowalski/socgo/internal/providers"
)

// WebHandler handles web page requests
type WebHandler struct {
	dbManager       *database.Manager
	providerService *providers.ProviderService
	templates       *template.Template
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
func NewWebHandler(dbManager *database.Manager, providerService *providers.ProviderService) *WebHandler {
	h := &WebHandler{
		dbManager:       dbManager,
		providerService: providerService,
	}
	h.loadTemplates()
	return h
}

func (h *WebHandler) loadTemplates() {
	var err error
	h.templates, err = template.ParseGlob("templates/*.tmpl")
	if err != nil {
		log.Printf("Error loading templates: %v", err)
		return
	}

	// Load partials
	partials, err := filepath.Glob("templates/partials/*.tmpl")
	if err == nil && len(partials) > 0 {
		h.templates, err = h.templates.ParseFiles(partials...)
		if err != nil {
			log.Printf("Error loading partial templates: %v", err)
		}
	}
}

// Helper method to render templates with flash message support
func (h *WebHandler) render(w http.ResponseWriter, r *http.Request, templateName string, data PageData) {
	// Check for flash messages in headers (from HTMX)
	if flashMsg := r.Header.Get("HX-Flash-Message"); flashMsg != "" {
		data.FlashMessage = flashMsg
		data.FlashType = r.Header.Get("HX-Flash-Type")
		if data.FlashType == "" {
			data.FlashType = "info"
		}
	}

	// Check for flash messages in query parameters (from redirects)
	if flashMsg := r.URL.Query().Get("flash"); flashMsg != "" {
		if decoded, err := url.QueryUnescape(flashMsg); err == nil {
			data.FlashMessage = decoded
		} else {
			data.FlashMessage = flashMsg
		}
		data.FlashType = r.URL.Query().Get("flash_type")
		if data.FlashType == "" {
			data.FlashType = "info"
		}
	}

	w.Header().Set("Content-Type", "text/html")
	if err := h.templates.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		log.Printf("Error executing template %s: %v", templateName, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
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
	data := PageData{
		Title:       "Home",
		CurrentPage: "home",
	}
	h.render(w, r, "home.tmpl", data)
}

// DashboardPage handles the dashboard page
func (h *WebHandler) DashboardPage(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:       "Dashboard",
		CurrentPage: "dashboard",
	}
	h.render(w, r, "dashboard.tmpl", data)
}

// ProvidersPage handles the providers page
func (h *WebHandler) ProvidersPage(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:       "Providers",
		CurrentPage: "providers",
	}
	h.render(w, r, "providers.tmpl", data)
}

// PostsPage handles the posts page
func (h *WebHandler) PostsPage(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:       "Create Post",
		CurrentPage: "posts",
	}
	h.render(w, r, "posts.tmpl", data)
}

// CalendarPage handles the calendar page
func (h *WebHandler) CalendarPage(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:       "Calendar",
		CurrentPage: "calendar",
	}
	h.render(w, r, "calendar.tmpl", data)
}

// HandlePost handles form submissions for creating posts
func (h *WebHandler) HandlePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.setFlashMessage(w, "Method not allowed", "error")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		h.setFlashMessage(w, "Invalid form data", "error")
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	providerID := r.FormValue("provider_id")
	content := strings.TrimSpace(r.FormValue("content"))
	scheduleType := r.FormValue("schedule_type")
	scheduleAt := r.FormValue("schedule_at")

	// Basic validation
	if providerID == "" {
		h.setFlashMessage(w, "Please select a provider", "error")
		http.Error(w, "Provider is required", http.StatusBadRequest)
		return
	}
	if content == "" {
		h.setFlashMessage(w, "Content is required", "error")
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	// Convert to JSON request for existing handler
	var req PostRequest
	if err := json.Unmarshal([]byte(fmt.Sprintf(`{
		"provider_id": %s,
		"content": %q,
		"schedule_at": %q
	}`, providerID, content, func() string {
		if scheduleType == "scheduled" && scheduleAt != "" {
			// Convert HTML datetime-local format to RFC3339
			if t, err := time.Parse("2006-01-02T15:04", scheduleAt); err == nil {
				return t.Format(time.RFC3339)
			}
		}
		return "now"
	}())), &req); err != nil {
		h.setFlashMessage(w, "Invalid request format", "error")
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Use existing post handler logic
	postHandler := NewPostHandler(h.dbManager, h.providerService)
	
	// Create a custom response writer to capture the response
	responseCapture := &responseCapture{ResponseWriter: w}
	
	// Create new request with JSON body
	reqJSON := strings.NewReader(fmt.Sprintf(`{
		"provider_id": %s,
		"content": %q,
		"schedule_at": %q
	}`, providerID, content, func() string {
		if scheduleType == "scheduled" && scheduleAt != "" {
			if t, err := time.Parse("2006-01-02T15:04", scheduleAt); err == nil {
				return t.Format(time.RFC3339)
			}
		}
		return "now"
	}()))
	
	newReq := r.Clone(r.Context())
	newReq.Body = io.NopCloser(reqJSON)
	newReq.Header.Set("Content-Type", "application/json")
	
	// Call the existing handler
	postHandler.HandlePost(responseCapture, newReq)
	
	// Handle response based on status
	if responseCapture.statusCode >= 200 && responseCapture.statusCode < 300 {
		var response PostResponse
		message := "Post created successfully"
		if err := json.Unmarshal(responseCapture.body, &response); err == nil {
			message = response.Message
		}
		
		// Check if this is an HTMX request
		if r.Header.Get("HX-Request") == "true" {
			h.setFlashMessage(w, message, "success")
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`<div class="p-4 bg-green-100 text-green-800 rounded-lg">✓ Post created successfully!</div>`)); err != nil {
				log.Printf("Error writing success response: %v", err)
			}
		} else {
			// Regular form submission - redirect with flash message
			h.redirectWithFlash(w, r, "/posts", message, "success")
		}
	} else {
		message := "Failed to create post"
		
		// Check if this is an HTMX request
		if r.Header.Get("HX-Request") == "true" {
			h.setFlashMessage(w, message, "error")
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(responseCapture.statusCode)
			if _, err := w.Write([]byte(`<div class="p-4 bg-red-100 text-red-800 rounded-lg">❌ Failed to create post. Please try again.</div>`)); err != nil {
				log.Printf("Error writing error response: %v", err)
			}
		} else {
			// Regular form submission - redirect with flash message
			h.redirectWithFlash(w, r, "/posts", message, "error")
		}
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
	db.Model(&database.Provider{}).Count(&count)
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
	db.Model(&database.Post{}).Where("user_id = ?", userID).Count(&count)
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
	db.Model(&database.ScheduledJob{}).Where("user_id = ? AND status = ?", userID, "pending").Count(&count)
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
	db.Model(&database.Post{}).Where("user_id = ? AND created_at >= ? AND created_at <= ?", 
		userID, startOfMonth, endOfMonth).Count(&count)
	if _, err := w.Write([]byte(fmt.Sprintf("%d", count))); err != nil {
		log.Printf("Error writing monthly count: %v", err)
	}
}

// HandleProvidersOptions returns HTML options for provider select
func (h *WebHandler) HandleProvidersOptions(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	db, err := h.dbManager.GetDB(userID)
	if err != nil {
		if _, writeErr := w.Write([]byte(`<option value="">Error loading providers</option>`)); writeErr != nil {
			log.Printf("Error writing response: %v", writeErr)
		}
		return
	}

	var providers []database.Provider
	if err := db.Find(&providers).Error; err != nil {
		if _, writeErr := w.Write([]byte(`<option value="">Error loading providers</option>`)); writeErr != nil {
			log.Printf("Error writing response: %v", writeErr)
		}
		return
	}

	html := `<option value="">Choose a provider...</option>`
	for _, provider := range providers {
		// Check if provider is configured
		configured, err := h.providerService.IsProviderConfigured(userID, provider.Name)
		if err != nil || !configured {
			html += fmt.Sprintf(`<option value="%d" disabled>%s (not connected)</option>`, provider.ID, provider.Name)
		} else {
			html += fmt.Sprintf(`<option value="%d">%s</option>`, provider.ID, provider.Name)
		}
	}

	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(html)); err != nil {
		log.Printf("Error writing providers options: %v", err)
	}
}

func (h *WebHandler) getUserID(r *http.Request) string {
	// TODO: Implement proper user authentication
	return "default_user"
}