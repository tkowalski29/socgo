package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tkowalski/socgo/internal/service/notifications"
)

type NotificationHandler struct {
	notificationService *notifications.Service
}

func NewNotificationHandler(notificationService *notifications.Service) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// HandleGetNotificationGroups returns grouped notifications
func (h *NotificationHandler) HandleGetNotificationGroups(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromRequest(r)
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
	json.NewEncoder(w).Encode(groups)
}

// HandleGetNotificationsByGroup returns all notifications in a group
func (h *NotificationHandler) HandleGetNotificationsByGroup(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromRequest(r)
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
	json.NewEncoder(w).Encode(notifications)
}

// HandleMarkGroupAsRead marks all notifications in a group as read
func (h *NotificationHandler) HandleMarkGroupAsRead(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromRequest(r)
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
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// HandleGetNotificationStats returns notification statistics
func (h *NotificationHandler) HandleGetNotificationStats(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromRequest(r)
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
	json.NewEncoder(w).Encode(stats)
}

// getUserIDFromRequest extracts user ID from request
func getUserIDFromRequest(r *http.Request) string {
	// Brak logowania – zawsze domyślny user
	return "default_user"
}
