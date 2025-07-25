package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/tkowalski/socgo/internal/handlers/internal"
	"github.com/tkowalski/socgo/internal/service/notifications"
	"github.com/tkowalski/socgo/web/templates/component"
)

type NotificationHandler struct {
	notificationService *notifications.Service
}

func NewNotificationHandler(notificationService *notifications.Service) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// HandleNotificationGroups handles notification groups requests
func (h *NotificationHandler) HandleNotificationGroups(w http.ResponseWriter, r *http.Request) {
	userID := internal.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if h.notificationService == nil {
		// Return empty groups list if notification service is not available
		var buf strings.Builder
		if err := component.NotificationGroupsList([]component.NotificationGroupData{}).Render(r.Context(), &buf); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		if _, err := w.Write([]byte(buf.String())); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		return
	}

	notificationType := r.URL.Query().Get("type")
	groups, err := h.notificationService.GetNotificationGroups(userID, notificationType)
	if err != nil {
		// Return empty groups list on error
		var buf strings.Builder
		if err := component.NotificationGroupsList([]component.NotificationGroupData{}).Render(r.Context(), &buf); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		if _, err := w.Write([]byte(buf.String())); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Convert to component data
	var groupData []component.NotificationGroupData
	for _, group := range groups {
		groupData = append(groupData, component.NotificationGroupData{
			GroupID:       group.GroupID,
			Type:          group.Type,
			Title:         group.Title,
			LatestMessage: group.LatestMessage,
			Count:         group.Count,
			UnreadCount:   group.UnreadCount,
			IsActive:      group.IsActive,
			LatestAt:      group.LatestAt,
			PostID:        group.PostID,
		})
	}

	var buf strings.Builder
	if err := component.NotificationGroupsList(groupData).Render(r.Context(), &buf); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(buf.String())); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// HandleNotificationDetails handles notification details requests
func (h *NotificationHandler) HandleNotificationDetails(w http.ResponseWriter, r *http.Request) {
	userID := internal.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	groupID := vars["groupID"]
	if groupID == "" {
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	notifications, err := h.notificationService.GetNotificationsByGroup(userID, groupID)
	if err != nil {
		http.Error(w, "Failed to get notification details", http.StatusInternalServerError)
		return
	}

	// Convert to component data
	var notificationData []component.NotificationData
	for _, notification := range notifications {
		notificationData = append(notificationData, component.NotificationData{
			ID:        notification.ID,
			UserID:    notification.UserID,
			Type:      notification.Type,
			Category:  notification.Category,
			Title:     notification.Title,
			Message:   notification.Message,
			PostID:    notification.PostID,
			GroupID:   notification.GroupID,
			IsRead:    notification.IsRead,
			IsActive:  notification.IsActive,
			CreatedAt: notification.CreatedAt,
			UpdatedAt: notification.UpdatedAt,
		})
	}

	currentTab := r.URL.Query().Get("tab")
	if currentTab == "" {
		currentTab = "all"
	}

	var buf strings.Builder
	if err := component.NotificationDetailsList(notificationData, currentTab).Render(r.Context(), &buf); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(buf.String())); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// HandleNotificationBell handles notification bell requests
func (h *NotificationHandler) HandleNotificationBell(w http.ResponseWriter, r *http.Request) {
	userID := internal.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if h.notificationService == nil {
		// Return empty bell if notification service is not available
		var buf strings.Builder
		if err := component.NotificationBellIcon(0).Render(r.Context(), &buf); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		if _, err := w.Write([]byte(buf.String())); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		return
	}

	stats, err := h.notificationService.GetNotificationStats(userID)
	if err != nil {
		// Return empty bell on error
		var buf strings.Builder
		if err := component.NotificationBellIcon(0).Render(r.Context(), &buf); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		if _, err := w.Write([]byte(buf.String())); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		return
	}

	var buf strings.Builder
	if err := component.NotificationBellIcon(stats.UnreadCount).Render(r.Context(), &buf); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(buf.String())); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// HandleGetNotificationGroups returns grouped notifications
func (h *NotificationHandler) HandleGetNotificationGroups(w http.ResponseWriter, r *http.Request) {
	userID := internal.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	notificationType := r.URL.Query().Get("type")

	groups, err := h.notificationService.GetNotificationGroups(userID, notificationType)
	if err != nil {
		http.Error(w, "Failed to get notifications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(groups); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// HandleGetNotificationsByGroup returns all notifications in a group
func (h *NotificationHandler) HandleGetNotificationsByGroup(w http.ResponseWriter, r *http.Request) {
	userID := internal.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	groupID := vars["groupID"]
	if groupID == "" {
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	notifications, err := h.notificationService.GetNotificationsByGroup(userID, groupID)
	if err != nil {
		http.Error(w, "Failed to get notifications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(notifications); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// HandleMarkGroupAsRead marks all notifications in a group as read
func (h *NotificationHandler) HandleMarkGroupAsRead(w http.ResponseWriter, r *http.Request) {
	userID := internal.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	groupID := vars["groupID"]
	if groupID == "" {
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	err := h.notificationService.MarkGroupAsRead(userID, groupID)
	if err != nil {
		http.Error(w, "Failed to mark notifications as read", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "success"}); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// HandleGetNotificationStats returns notification statistics
func (h *NotificationHandler) HandleGetNotificationStats(w http.ResponseWriter, r *http.Request) {
	userID := internal.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats, err := h.notificationService.GetNotificationStats(userID)
	if err != nil {
		http.Error(w, "Failed to get notification stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
