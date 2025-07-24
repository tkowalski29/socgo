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

	data_database "github.com/tkowalski/socgo/internal/data/database"
	provider_pkg "github.com/tkowalski/socgo/internal/data/provider"
	"github.com/tkowalski/socgo/internal/service/database"
	"github.com/tkowalski/socgo/internal/service/post"
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

type CalendarDay struct {
	Day       int  `json:"day"`
	HasPosts  bool `json:"has_posts"`
	PostCount int  `json:"post_count"`
}

type CalendarResponse struct {
	Year  int           `json:"year"`
	Month int           `json:"month"`
	Days  []CalendarDay `json:"days"`
}

type WeekPost struct {
	ID           uint                   `json:"id"`
	Content      string                 `json:"content"`
	ProviderID   uint                   `json:"provider_id"`
	Provider     data_database.Provider `json:"provider"`
	ScheduledAt  *time.Time             `json:"scheduled_at,omitempty"`
	PublishedAt  *time.Time             `json:"published_at,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	Status       string                 `json:"status"`
	ExternalID   string                 `json:"external_id"`
	ExternalURL  string                 `json:"external_url"`
	ErrorMessage string                 `json:"error_message"`
	Type         string                 `json:"type"` // "scheduled" or "published"
	Hour         int                    `json:"hour"` // Hour of the day (0-23)
}

type WeekDay struct {
	Date  time.Time          `json:"date"`
	Day   int                `json:"day"`
	Hours map[int][]WeekPost `json:"hours"` // Map of hour -> posts
}

type WeekResponse struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Days      []WeekDay `json:"days"`
}

// PostHandler handles POST requests for creating posts
type PostHandler struct {
	dbManager       *database.Manager
	providerService *post.ProviderService
}

// NewPostHandler creates a new PostHandler instance
func NewPostHandler(dbManager *database.Manager, providerService *post.ProviderService) *PostHandler {
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
	var provider data_database.Provider
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
		postID, err := h.providerService.PublishContent(ctx, userID, provider.Name, req.Content, []provider_pkg.Media{}, nil)
		if err != nil {
			log.Printf("Error publishing content: %v", err)
			http.Error(w, "Failed to publish content", http.StatusInternalServerError)
			return
		}

		// Save post to database
		post := data_database.Post{
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
		job := data_database.ScheduledJob{
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

	if err := db.Model(&data_database.Post{}).
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

	if err := db.Model(&data_database.ScheduledJob{}).
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

func (h *PostHandler) HandleWeekView(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get start date parameter
	startDateStr := r.URL.Query().Get("start")

	now := time.Now()
	startDate := h.getStartOfWeek(now)

	if startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = h.getStartOfWeek(parsed)
		}
	}

	endDate := startDate.AddDate(0, 0, 7)

	userID := h.getUserID(r)
	db, err := h.dbManager.GetDB(userID)
	if err != nil {
		log.Printf("Error getting database: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get published posts for the week
	var publishedPosts []data_database.Post
	if err := db.Preload("Provider").
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startDate, endDate).
		Order("created_at ASC").
		Find(&publishedPosts).Error; err != nil {
		log.Printf("Error fetching published posts: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get scheduled posts for the week
	var scheduledJobs []data_database.ScheduledJob
	if err := db.Preload("Provider").
		Where("user_id = ? AND scheduled_at >= ? AND scheduled_at < ?", userID, startDate, endDate).
		Order("scheduled_at ASC").
		Find(&scheduledJobs).Error; err != nil {
		log.Printf("Error fetching scheduled jobs: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create week days
	days := make([]WeekDay, 7)
	for i := 0; i < 7; i++ {
		dayDate := startDate.AddDate(0, 0, i)
		days[i] = WeekDay{
			Date:  dayDate,
			Day:   dayDate.Day(),
			Hours: make(map[int][]WeekPost),
		}
	}

	// Add published posts to appropriate days and hours
	for _, post := range publishedPosts {
		// Calculate day index more accurately
		postDate := post.CreatedAt.Truncate(24 * time.Hour)
		startDateTruncated := startDate.Truncate(24 * time.Hour)
		dayIndex := int(postDate.Sub(startDateTruncated).Hours() / 24)

		if dayIndex >= 0 && dayIndex < 7 {
			hour := post.CreatedAt.Hour()
			days[dayIndex].Hours[hour] = append(days[dayIndex].Hours[hour], WeekPost{
				ID:           post.ID,
				Content:      post.Content,
				ProviderID:   post.ProviderID,
				Provider:     post.Provider,
				PublishedAt:  post.PublishedAt,
				CreatedAt:    post.CreatedAt,
				Status:       post.Status,
				ExternalID:   post.ExternalID,
				ExternalURL:  post.ExternalURL,
				ErrorMessage: post.ErrorMessage,
				Type:         "published",
				Hour:         hour,
			})
		}
	}

	// Add scheduled posts to appropriate days and hours
	for _, job := range scheduledJobs {
		// Calculate day index more accurately
		jobDate := job.ScheduledAt.Truncate(24 * time.Hour)
		startDateTruncated := startDate.Truncate(24 * time.Hour)
		dayIndex := int(jobDate.Sub(startDateTruncated).Hours() / 24)

		if dayIndex >= 0 && dayIndex < 7 {
			hour := job.ScheduledAt.Hour()
			days[dayIndex].Hours[hour] = append(days[dayIndex].Hours[hour], WeekPost{
				ID:          job.ID,
				Content:     job.PayloadData,
				ProviderID:  job.ProviderID,
				Provider:    job.Provider,
				ScheduledAt: &job.ScheduledAt,
				CreatedAt:   job.CreatedAt,
				Status:      job.Status,
				Type:        "scheduled",
				Hour:        hour,
			})
		}
	}

	// Check if request wants JSON (API) or HTML (HTMX)
	if r.Header.Get("Accept") == "application/json" {
		response := WeekResponse{
			StartDate: startDate,
			EndDate:   endDate,
			Days:      days,
		}
		h.writeJSONResponse(w, response, http.StatusOK)
		return
	}

	// Return HTML week view for HTMX
	h.renderWeekViewHTML(w, days, startDate)
}

func (h *PostHandler) renderWeekViewHTML(w http.ResponseWriter, days []WeekDay, startDate time.Time) {
	var htmlBuilder strings.Builder

	// Days of week header - Monday to Sunday
	daysOfWeek := []string{"Poniedziałek", "Wtorek", "Środa", "Czwartek", "Piątek", "Sobota", "Niedziela"}

	htmlBuilder.WriteString(`<div class="border border-gray-200 rounded-lg overflow-hidden w-full">`)

	// Header row with day names
	htmlBuilder.WriteString(`<div class="grid grid-cols-8 bg-gray-50 border-b border-gray-200">`)
	htmlBuilder.WriteString(`<div class="p-3 border-r border-gray-200"></div>`) // Empty corner cell

	for i, dayName := range daysOfWeek {
		dayDate := startDate.AddDate(0, 0, i)
		isToday := time.Now().Format("2006-01-02") == dayDate.Format("2006-01-02")

		dayClass := "p-3 text-center border-r border-gray-200"
		if isToday {
			dayClass += " bg-blue-50"
		}

		textColor := "text-gray-900"
		if isToday {
			textColor = "text-blue-600"
		}

		htmlBuilder.WriteString(fmt.Sprintf(`
			<div class="%s">
				<div class="text-sm font-medium text-gray-600">%s</div>
				<div class="text-lg font-semibold %s">%d</div>
			</div>
		`, dayClass, dayName, textColor, dayDate.Day()))
	}
	htmlBuilder.WriteString(`</div>`)

	// Get current time for future hour detection
	now := time.Now()

	// Table body with hours
	for hour := 0; hour < 24; hour++ {
		htmlBuilder.WriteString(`<div class="grid grid-cols-8 border-b border-gray-200 last:border-b-0">`)

		// Hour label on the left
		htmlBuilder.WriteString(fmt.Sprintf(`
			<div class="p-2 bg-gray-50 border-r border-gray-200 text-sm font-medium text-gray-700 text-center">
				%02d:00
			</div>
		`, hour))

		// Day cells for this hour
		for i, day := range days {
			dayDate := startDate.AddDate(0, 0, i)
			isToday := time.Now().Format("2006-01-02") == dayDate.Format("2006-01-02")

			// Check if this hour is in the future
			isFuture := false
			if isToday {
				currentHour := now.Hour()
				isFuture = hour > currentHour
			} else if dayDate.After(now) {
				isFuture = true
			}

			cellClass := "p-2 border-r border-gray-200 min-h-16 relative"
			if i == 6 { // Last column
				cellClass = "p-2 min-h-16 relative"
			}
			if isToday {
				cellClass += " bg-blue-50"
			}
			if isFuture {
				cellClass += " bg-green-50 hover:bg-green-100 transition-colors"
			}

			htmlBuilder.WriteString(fmt.Sprintf(`<div class="%s">`, cellClass))

			posts, exists := day.Hours[hour]
			if !exists || len(posts) == 0 {
				if isFuture {
					// Show + icon for future empty cells
					htmlBuilder.WriteString(fmt.Sprintf(`
						<div class="text-gray-300 text-xs text-center py-4 group cursor-pointer" onclick="openPostSidebarWithDateTime('%s', %d)">
							<div class="opacity-0 group-hover:opacity-100 transition-opacity">
								<svg class="w-6 h-6 mx-auto text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"></path>
								</svg>
								<div class="text-xs text-green-600 mt-1">Dodaj post</div>
							</div>
							<div class="group-hover:opacity-0 transition-opacity">-</div>
						</div>
					`, dayDate.Format("2006-01-02"), hour))
				} else {
					htmlBuilder.WriteString(`<div class="text-gray-300 text-xs text-center py-4">-</div>`)
				}
			} else {
				// Show first 2 posts
				visiblePosts := posts
				hiddenPosts := []WeekPost{}

				if len(posts) > 2 {
					visiblePosts = posts[:2]
					hiddenPosts = posts[2:]
				}

				// Render visible posts
				for _, post := range visiblePosts {
					statusClass := "bg-green-100 text-green-800"
					if post.Status == "pending" {
						statusClass = "bg-yellow-100 text-yellow-800"
					} else if post.Status == "failed" {
						statusClass = "bg-red-100 text-red-800"
					}

					timeStr := ""
					if post.ScheduledAt != nil {
						timeStr = post.ScheduledAt.Format("15:04")
					} else if post.PublishedAt != nil {
						timeStr = post.PublishedAt.Format("15:04")
					}

					htmlBuilder.WriteString(fmt.Sprintf(`
						<div class="mb-1 p-1 bg-white border border-gray-200 rounded text-xs shadow-sm hover:shadow-md transition-shadow cursor-pointer"
							 onclick="showCalendarPostSidebar(%d, '%s')">
							<div class="flex justify-between items-start mb-1">
								<span class="px-1 py-0.5 text-xs rounded %s">%s</span>
								<span class="text-xs text-gray-500">%s</span>
							</div>
							<div class="text-xs text-gray-800 line-clamp-1">%s</div>
							<div class="text-xs text-gray-500">%s</div>
						</div>
					`, post.ID, post.Type, statusClass, post.Status, timeStr,
						h.truncateText(post.Content, 40), post.Provider.Name))
				}

				// Show "..." if there are hidden posts
				if len(hiddenPosts) > 0 {
					htmlBuilder.WriteString(fmt.Sprintf(`
						<button 
							id="toggle-%d-%d"
							data-count="%d"
							onclick="toggleHourPosts(%d, %d)"
							class="w-full mt-1 px-1 py-0.5 text-xs bg-gray-100 hover:bg-gray-200 text-gray-700 rounded transition-colors"
						>
							+%d więcej
						</button>
						<div id="hidden-posts-%d-%d" class="hidden">
					`, i, hour, len(hiddenPosts), i, hour, len(hiddenPosts), i, hour))

					// Render hidden posts
					for _, post := range hiddenPosts {
						statusClass := "bg-green-100 text-green-800"
						if post.Status == "pending" {
							statusClass = "bg-yellow-100 text-yellow-800"
						} else if post.Status == "failed" {
							statusClass = "bg-red-100 text-red-800"
						}

						timeStr := ""
						if post.ScheduledAt != nil {
							timeStr = post.ScheduledAt.Format("15:04")
						} else if post.PublishedAt != nil {
							timeStr = post.PublishedAt.Format("15:04")
						}

						htmlBuilder.WriteString(fmt.Sprintf(`
							<div class="mb-1 p-1 bg-white border border-gray-200 rounded text-xs shadow-sm hover:shadow-md transition-shadow cursor-pointer"
								 onclick="showCalendarPostSidebar(%d, '%s')">
								<div class="flex justify-between items-start mb-1">
									<span class="px-1 py-0.5 text-xs rounded %s">%s</span>
									<span class="text-xs text-gray-500">%s</span>
								</div>
								<div class="text-xs text-gray-800 line-clamp-1">%s</div>
								<div class="text-xs text-gray-500">%s</div>
							</div>
						`, post.ID, post.Type, statusClass, post.Status, timeStr,
							h.truncateText(post.Content, 40), post.Provider.Name))
					}

					htmlBuilder.WriteString(`</div>`)
				}
			}

			htmlBuilder.WriteString(`</div>`)
		}

		htmlBuilder.WriteString(`</div>`)
	}

	htmlBuilder.WriteString(`</div>`)

	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(htmlBuilder.String())); err != nil {
		log.Printf("Error writing week view response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *PostHandler) getStartOfWeek(date time.Time) time.Time {
	weekday := int(date.Weekday())
	if weekday == 0 { // Sunday
		weekday = 7
	}
	// Start from Monday (weekday 1)
	return date.AddDate(0, 0, -weekday+1)
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

	userID := h.getUserID(r)
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
				
				<div class="flex space-x-3 pt-4 border-t">
					<button 
						onclick="editPost(%d, 'scheduled')"
						class="px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-lg transition-colors"
					>
						Edytuj
					</button>
					<button 
						onclick="deletePost(%d, 'scheduled')"
						class="px-4 py-2 bg-red-500 hover:bg-red-600 text-white rounded-lg transition-colors"
					>
						Usuń
					</button>
				</div>
			</div>
		`, job.ID, statusClass, job.Status, job.PayloadData, job.Provider.Name,
			job.ScheduledAt.Format("02.01.2006 15:04"), job.CreatedAt.Format("02.01.2006 15:04"),
			job.Status, job.ID, job.ID))

	} else {
		// Get published post with media
		var post data_database.Post
		if err := db.Preload("Provider").Preload("Media").First(&post, postID).Error; err != nil {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}

		statusClass := "bg-green-100 text-green-800"
		if post.Status == "pending" {
			statusClass = "bg-yellow-100 text-yellow-800"
		} else if post.Status == "failed" {
			statusClass = "bg-red-100 text-red-800"
		}

		htmlBuilder.WriteString(fmt.Sprintf(`
			<div class="space-y-4">
				<div class="flex justify-between items-start">
					<div>
						<h4 class="text-lg font-semibold text-gray-900">Opublikowany post</h4>
						<p class="text-sm text-gray-600">ID: %d</p>
					</div>
					<span class="px-3 py-1 text-sm rounded-full %s">%s</span>
				</div>
				
				<div class="bg-gray-50 p-4 rounded-lg">
					<h5 class="font-medium text-gray-900 mb-2">Treść:</h5>
					<p class="text-gray-800 whitespace-pre-wrap">%s</p>
				</div>
		`, post.ID, statusClass, post.Status, post.Content))

		// Add media section if there are media files
		if len(post.Media) > 0 {
			htmlBuilder.WriteString(`
				<div class="bg-blue-50 p-4 rounded-lg">
					<h5 class="font-medium text-gray-900 mb-3">Załączone media:</h5>
					<div class="grid grid-cols-4 gap-2">
			`)

			for _, media := range post.Media {
				// Determine the correct file path for display
				displayPath := media.FileName
				if strings.Contains(media.FilePath, "/") {
					// Old format: full path stored in FilePath, use FileName for display
					displayPath = media.FileName
				} else {
					// New format: FileName contains UUID filename
					displayPath = media.FileName
				}

				// Check if it's an image by FileType or file extension
				isImage := strings.HasPrefix(media.FileType, "image/") ||
					media.FileType == "image" ||
					strings.HasSuffix(strings.ToLower(displayPath), ".jpg") ||
					strings.HasSuffix(strings.ToLower(displayPath), ".jpeg") ||
					strings.HasSuffix(strings.ToLower(displayPath), ".png") ||
					strings.HasSuffix(strings.ToLower(displayPath), ".gif") ||
					strings.HasSuffix(strings.ToLower(displayPath), ".webp")

				if isImage {
					// Small image thumbnail
					htmlBuilder.WriteString(fmt.Sprintf(`
						<div class="relative group cursor-pointer" onclick="window.open('/uploads/%s', '_blank')">
							<img src="/uploads/%s" alt="Media" class="w-full h-16 object-cover rounded-md border border-gray-200 shadow-sm hover:shadow-md transition-shadow">
							<div class="absolute inset-0 bg-black bg-opacity-0 group-hover:bg-opacity-10 transition-all rounded-md"></div>
						</div>
					`, displayPath, displayPath))
				} else {
					// Small file icon
					htmlBuilder.WriteString(fmt.Sprintf(`
						<div class="bg-white p-2 rounded-md border border-gray-200 shadow-sm hover:shadow-md transition-shadow cursor-pointer" onclick="window.open('/uploads/%s', '_blank')">
							<div class="flex flex-col items-center space-y-1">
								<svg class="w-6 h-6 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
								</svg>
								<span class="text-xs text-gray-600 truncate w-full text-center">%s</span>
							</div>
						</div>
					`, displayPath, displayPath))
				}
			}

			htmlBuilder.WriteString(`
					</div>
				</div>
			`)
		}

		// Add provider information
		htmlBuilder.WriteString(fmt.Sprintf(`
				<div class="bg-gray-50 p-4 rounded-lg">
					<h5 class="font-medium text-gray-900 mb-3">Informacje o platformie:</h5>
					<div class="grid grid-cols-1 gap-3">
						<div class="flex items-center space-x-3">
							<div class="w-8 h-8 bg-blue-100 rounded-lg flex items-center justify-center">
								<svg class="w-5 h-5 text-blue-600" fill="currentColor" viewBox="0 0 24 24">
									<path d="M24 4.557c-.883.392-1.832.656-2.828.775 1.017-.609 1.798-1.574 2.165-2.724-.951.564-2.005.974-3.127 1.195-.897-.957-2.178-1.555-3.594-1.555-3.179 0-5.515 2.966-4.797 6.045-4.091-.205-7.719-2.165-10.148-5.144-1.29 2.213-.669 5.108 1.523 6.574-.806-.026-1.566-.247-2.229-.616-.054 2.281 1.581 4.415 3.949 4.89-.693.188-1.452.232-2.224.084.626 1.956 2.444 3.379 4.6 3.419-2.07 1.623-4.678 2.348-7.29 2.04 2.179 1.397 4.768 2.212 7.548 2.212 9.142 0 14.307-7.721 13.995-14.646.962-.695 1.797-1.562 2.457-2.549z"/>
								</svg>
							</div>
							<div>
								<div class="font-medium text-gray-900">%s</div>
								<div class="text-sm text-gray-600">%s</div>
							</div>
						</div>
					</div>
				</div>
		`, post.Provider.Name, strings.Title(post.Provider.Type)))

		htmlBuilder.WriteString(fmt.Sprintf(`
				<div class="grid grid-cols-2 gap-4">
					<div>
						<h5 class="font-medium text-gray-900 mb-1">Status:</h5>
						<p class="text-gray-600">%s</p>
					</div>
					<div>
						<h5 class="font-medium text-gray-900 mb-1">Utworzony:</h5>
						<p class="text-gray-600">%s</p>
					</div>
					<div>
						<h5 class="font-medium text-gray-900 mb-1">Opublikowany:</h5>
						<p class="text-gray-600">%s</p>
					</div>
					<div>
						<h5 class="font-medium text-gray-900 mb-1">Typ:</h5>
						<p class="text-gray-600">%s</p>
					</div>
				</div>
		`, post.Status, post.CreatedAt.Format("02.01.2006 15:04"),
			h.formatPublishedAt(post.PublishedAt), strings.Title(post.Provider.Type)))

		// Add external links if available
		if post.ExternalID != "" || post.ExternalURL != "" {
			htmlBuilder.WriteString(`
				<div class="bg-blue-50 p-4 rounded-lg">
					<h5 class="font-medium text-gray-900 mb-2">Linki zewnętrzne:</h5>
			`)

			if post.ExternalID != "" {
				htmlBuilder.WriteString(fmt.Sprintf(`
					<div class="mb-2">
						<span class="text-sm font-medium text-gray-700">ID zewnętrzne:</span>
						<span class="text-sm text-gray-600 ml-2">%s</span>
					</div>
				`, post.ExternalID))
			}

			if post.ExternalURL != "" {
				htmlBuilder.WriteString(fmt.Sprintf(`
					<div>
						<span class="text-sm font-medium text-gray-700">URL:</span>
						<a href="%s" target="_blank" class="text-sm text-blue-600 hover:text-blue-800 ml-2 underline">Zobacz post</a>
					</div>
				`, post.ExternalURL))
			}

			htmlBuilder.WriteString(`</div>`)
		}

		// Add error message if failed
		if post.Status == "failed" && post.ErrorMessage != "" {
			htmlBuilder.WriteString(fmt.Sprintf(`
				<div class="bg-red-50 p-4 rounded-lg">
					<h5 class="font-medium text-red-900 mb-2">Błąd:</h5>
					<p class="text-red-800 text-sm">%s</p>
				</div>
			`, post.ErrorMessage))
		}

		htmlBuilder.WriteString(`</div>`)
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
		return "Nie opublikowano"
	}
	return publishedAt.Format("02.01.2006 15:04")
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

	userID := h.getUserID(r)
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

		// Only allow deletion of pending jobs
		if job.Status != "pending" {
			http.Error(w, "Can only delete pending scheduled posts", http.StatusBadRequest)
			return
		}

		if err := db.Delete(&job).Error; err != nil {
			log.Printf("Error deleting scheduled job: %v", err)
			http.Error(w, "Failed to delete scheduled post", http.StatusInternalServerError)
			return
		}

	} else {
		// Delete published post
		var post data_database.Post
		if err := db.First(&post, postID).Error; err != nil {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}

		// Only allow deletion of failed posts (published posts cannot be deleted)
		if post.Status != "failed" {
			http.Error(w, "Can only delete failed posts", http.StatusBadRequest)
			return
		}

		if err := db.Delete(&post).Error; err != nil {
			log.Printf("Error deleting post: %v", err)
			http.Error(w, "Failed to delete post", http.StatusInternalServerError)
			return
		}
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Post deleted successfully"}`))
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
