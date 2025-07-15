package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
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

type HistoryPost struct {
	ID          uint      `json:"id"`
	Content     string    `json:"content"`
	ProviderID  uint      `json:"provider_id"`
	Provider    database.Provider `json:"provider"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	Status      string    `json:"status"`
}

type HistoryResponse struct {
	Posts []HistoryPost `json:"posts"`
	Page  int           `json:"page"`
	Total int64         `json:"total"`
}

type CalendarDay struct {
	Day      int  `json:"day"`
	HasPosts bool `json:"has_posts"`
	PostCount int `json:"post_count"`
}

type CalendarResponse struct {
	Year  int           `json:"year"`
	Month int           `json:"month"`
	Days  []CalendarDay `json:"days"`
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

	userID := h.getUserID(r)
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
	if err := db.Model(&database.Post{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		log.Printf("Error counting posts: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get posts with pagination and include scheduled jobs
	var posts []database.Post
	if err := db.Preload("Provider").Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(pageSize).Offset(offset).
		Find(&posts).Error; err != nil {
		log.Printf("Error fetching posts: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get scheduled jobs
	var scheduledJobs []database.ScheduledJob
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
		historyPosts = append(historyPosts, HistoryPost{
			ID:          job.ID,
			Content:     job.PayloadData,
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
		statusClass := "bg-green-100 text-green-800"
		if post.Status == "pending" {
			statusClass = "bg-yellow-100 text-yellow-800"
		} else if post.Status == "failed" {
			statusClass = "bg-red-100 text-red-800"
		}

		scheduledText := ""
		if post.ScheduledAt != nil {
			scheduledText = fmt.Sprintf(" (scheduled for %s)", post.ScheduledAt.Format("Jan 02, 15:04"))
		}

		htmlBuilder.WriteString(fmt.Sprintf(`
			<div class="border rounded-lg p-4 bg-white">
				<div class="flex justify-between items-start mb-2">
					<span class="px-2 py-1 text-xs rounded %s">%s</span>
					<span class="text-sm text-gray-500">%s%s</span>
				</div>
				<p class="text-gray-800">%s</p>
				<div class="mt-2 text-xs text-gray-500">
					Provider: %s
				</div>
			</div>
		`, statusClass, post.Status, post.CreatedAt.Format("Jan 02, 15:04"), scheduledText, post.Content, post.Provider.Name))
	}
	
	htmlBuilder.WriteString(`</div>`)
	
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(htmlBuilder.String())); err != nil {
		log.Printf("Error writing history response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *PostHandler) HandleCalendar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get year and month parameters
	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")
	
	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}
	if monthStr != "" {
		if m, err := strconv.Atoi(monthStr); err == nil && m >= 1 && m <= 12 {
			month = m
		}
	}

	userID := h.getUserID(r)
	db, err := h.dbManager.GetDB(userID)
	if err != nil {
		log.Printf("Error getting database: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Calculate month boundaries
	startOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	// Get posts count by day for published posts
	var postCounts []struct {
		Day   int
		Count int
	}
	
	if err := db.Model(&database.Post{}).
		Select("EXTRACT(DAY FROM created_at) as day, COUNT(*) as count").
		Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, startOfMonth, endOfMonth).
		Group("EXTRACT(DAY FROM created_at)").
		Scan(&postCounts).Error; err != nil {
		log.Printf("Error fetching post counts: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get scheduled posts count by day
	var scheduledCounts []struct {
		Day   int
		Count int
	}
	
	if err := db.Model(&database.ScheduledJob{}).
		Select("EXTRACT(DAY FROM scheduled_at) as day, COUNT(*) as count").
		Where("user_id = ? AND scheduled_at >= ? AND scheduled_at <= ?", userID, startOfMonth, endOfMonth).
		Group("EXTRACT(DAY FROM scheduled_at)").
		Scan(&scheduledCounts).Error; err != nil {
		log.Printf("Error fetching scheduled counts: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create calendar days
	daysInMonth := endOfMonth.Day()
	days := make([]CalendarDay, daysInMonth)
	
	// Initialize all days
	for i := 0; i < daysInMonth; i++ {
		days[i] = CalendarDay{
			Day:       i + 1,
			HasPosts:  false,
			PostCount: 0,
		}
	}

	// Add post counts
	for _, pc := range postCounts {
		if pc.Day > 0 && pc.Day <= daysInMonth {
			days[pc.Day-1].PostCount += pc.Count
			days[pc.Day-1].HasPosts = true
		}
	}

	// Add scheduled counts
	for _, sc := range scheduledCounts {
		if sc.Day > 0 && sc.Day <= daysInMonth {
			days[sc.Day-1].PostCount += sc.Count
			days[sc.Day-1].HasPosts = true
		}
	}

	// Check if request wants JSON (API) or HTML (HTMX)
	if r.Header.Get("Accept") == "application/json" {
		response := CalendarResponse{
			Year:  year,
			Month: month,
			Days:  days,
		}
		h.writeJSONResponse(w, response, http.StatusOK)
		return
	}

	// Return HTML calendar grid for HTMX
	var htmlBuilder strings.Builder
	htmlBuilder.WriteString(`<div class="grid grid-cols-7 gap-1 text-center">`)
	
	// Days of week header
	daysOfWeek := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	for _, day := range daysOfWeek {
		htmlBuilder.WriteString(fmt.Sprintf(`<div class="font-semibold text-gray-600 p-2">%s</div>`, day))
	}

	// Get first day of month to calculate starting position
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	startingWeekday := int(firstDay.Weekday())

	// Add empty cells for days before month starts
	for i := 0; i < startingWeekday; i++ {
		htmlBuilder.WriteString(`<div class="p-2"></div>`)
	}

	// Add calendar days
	for _, day := range days {
		dayClass := "p-2 border border-gray-200 hover:bg-gray-100"
		if day.HasPosts {
			dayClass = "p-2 border border-blue-500 bg-blue-100 text-blue-800 font-semibold hover:bg-blue-200"
		}

		postCountText := ""
		if day.PostCount > 0 {
			postCountText = fmt.Sprintf(`<br><span class="text-xs">(%d posts)</span>`, day.PostCount)
		}

		htmlBuilder.WriteString(fmt.Sprintf(`
			<div class="%s">
				%d%s
			</div>
		`, dayClass, day.Day, postCountText))
	}
	
	htmlBuilder.WriteString(`</div>`)
	
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(htmlBuilder.String())); err != nil {
		log.Printf("Error writing calendar response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
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

var calendarTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Posts Calendar - SocGo</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body class="bg-gray-100">
    <div class="container mx-auto px-4 py-8">
        <h1 class="text-4xl font-bold text-center text-gray-800 mb-8">
            Posts Calendar & History
        </h1>
        
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
            <!-- Calendar Section -->
            <div class="bg-white rounded-lg shadow-md p-6">
                <h2 class="text-2xl font-bold text-gray-800 mb-4">Calendar</h2>
                <div class="mb-4 flex justify-between items-center">
                    <button 
                        hx-get="/posts/calendar?year={{.PrevYear}}&month={{.PrevMonth}}" 
                        hx-target="#calendar-grid"
                        class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"
                    >
                        ← Previous
                    </button>
                    <h3 class="text-xl font-semibold">{{.MonthName}} {{.Year}}</h3>
                    <button 
                        hx-get="/posts/calendar?year={{.NextYear}}&month={{.NextMonth}}" 
                        hx-target="#calendar-grid"
                        class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"
                    >
                        Next →
                    </button>
                </div>
                <div id="calendar-grid" 
                     hx-get="/posts/calendar" 
                     hx-trigger="load"
                     hx-target="this">
                    Loading calendar...
                </div>
            </div>

            <!-- History Section -->
            <div class="bg-white rounded-lg shadow-md p-6">
                <h2 class="text-2xl font-bold text-gray-800 mb-4">Recent Posts</h2>
                <div id="history-list" 
                     hx-get="/posts/history" 
                     hx-trigger="load"
                     hx-target="this">
                    Loading history...
                </div>
                <div class="mt-4 text-center">
                    <button 
                        hx-get="/posts/history?page=2" 
                        hx-target="#history-list"
                        class="bg-gray-500 hover:bg-gray-700 text-white font-bold py-2 px-4 rounded"
                    >
                        Load More
                    </button>
                </div>
            </div>
        </div>
    </div>
</body>
</html>
`

func (h *PostHandler) HandleCalendarPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("calendar").Parse(calendarTemplate)
	if err != nil {
		log.Printf("Error parsing calendar template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	// Get year and month from query parameters
	if yearStr := r.URL.Query().Get("year"); yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}
	if monthStr := r.URL.Query().Get("month"); monthStr != "" {
		if m, err := strconv.Atoi(monthStr); err == nil && m >= 1 && m <= 12 {
			month = m
		}
	}

	// Calculate previous and next month/year
	prevMonth := month - 1
	prevYear := year
	if prevMonth < 1 {
		prevMonth = 12
		prevYear--
	}

	nextMonth := month + 1
	nextYear := year
	if nextMonth > 12 {
		nextMonth = 1
		nextYear++
	}

	data := struct {
		Year      int
		Month     int
		MonthName string
		PrevYear  int
		PrevMonth int
		NextYear  int
		NextMonth int
	}{
		Year:      year,
		Month:     month,
		MonthName: time.Month(month).String(),
		PrevYear:  prevYear,
		PrevMonth: prevMonth,
		NextYear:  nextYear,
		NextMonth: nextMonth,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing calendar template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

