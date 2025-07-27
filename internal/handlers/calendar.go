package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	data_database "github.com/tkowalski/socgo/internal/data/database"
	"github.com/tkowalski/socgo/internal/handlers/internal"
	"github.com/tkowalski/socgo/internal/service/database"
	"github.com/tkowalski/socgo/internal/service/post"
	"github.com/tkowalski/socgo/web/templates"
)

// Calendar-related data structures
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

// CalendarHandler handles calendar-related requests
type CalendarHandler struct {
	dbManager       *database.Manager
	providerService *post.ProviderService
}

// NewCalendarHandler creates a new CalendarHandler instance
func NewCalendarHandler(dbManager *database.Manager, providerService *post.ProviderService) *CalendarHandler {
	return &CalendarHandler{
		dbManager:       dbManager,
		providerService: providerService,
	}
}

// CalendarPage handles the calendar page
func (h *CalendarHandler) CalendarPage(w http.ResponseWriter, r *http.Request) {
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

func (h *CalendarHandler) HandleCalendar(w http.ResponseWriter, r *http.Request) {
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

	userID := internal.GetUserID(r)
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

func (h *CalendarHandler) HandleWeekView(w http.ResponseWriter, r *http.Request) {
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

	userID := internal.GetUserID(r)
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
		// Parse PayloadData to extract content
		var payloadData struct {
			Content string `json:"content"`
		}
		
		// Try to parse as JSON first, if it fails, treat as plain text
		if err := json.Unmarshal([]byte(job.PayloadData), &payloadData); err != nil {
			// If it's not JSON, treat as plain text content
			payloadData.Content = job.PayloadData
		}
		
		// Calculate day index more accurately
		jobDate := job.ScheduledAt.Truncate(24 * time.Hour)
		startDateTruncated := startDate.Truncate(24 * time.Hour)
		dayIndex := int(jobDate.Sub(startDateTruncated).Hours() / 24)

		if dayIndex >= 0 && dayIndex < 7 {
			hour := job.ScheduledAt.Hour()
			days[dayIndex].Hours[hour] = append(days[dayIndex].Hours[hour], WeekPost{
				ID:          job.ID,
				Content:     payloadData.Content,
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

func (h *CalendarHandler) renderWeekViewHTML(w http.ResponseWriter, days []WeekDay, startDate time.Time) {
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

func (h *CalendarHandler) getStartOfWeek(date time.Time) time.Time {
	weekday := int(date.Weekday())
	if weekday == 0 { // Sunday
		weekday = 7
	}
	// Start from Monday (weekday 1)
	return date.AddDate(0, 0, -weekday+1)
}

func (h *CalendarHandler) truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength-3] + "..."
}

func (h *CalendarHandler) writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}
