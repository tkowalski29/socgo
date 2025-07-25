package database

import (
	"time"
)

// Notification represents a notification in the system
type Notification struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	UserID    string    `gorm:"index;not null"`
	Type      string    `gorm:"not null"` // "app", "post", "schedule"
	Category  string    `gorm:"not null"` // "error", "success", "warning", "info"
	Title     string    `gorm:"not null"`
	Message   string    `gorm:"not null"`
	PostID    *uint     `gorm:"index"`          // Optional, for post-related notifications
	GroupID   string    `gorm:"index;not null"` // Groups notifications by post or action
	IsRead    bool      `gorm:"default:false"`
	IsActive  bool      `gorm:"default:true"` // For reactivating groups
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// NotificationGroup represents a group of notifications
type NotificationGroup struct {
	GroupID       string    `json:"group_id" gorm:"column:group_id"`
	Type          string    `json:"type" gorm:"column:type"`
	Title         string    `json:"title" gorm:"column:title"`
	LatestMessage string    `json:"latest_message" gorm:"column:latest_message"`
	Count         int       `json:"count" gorm:"column:count"`
	UnreadCount   int       `json:"unread_count" gorm:"column:unread_count"`
	IsActive      bool      `json:"is_active" gorm:"column:is_active"`
	LatestAt      time.Time `json:"latest_at" gorm:"column:latest_at"`
	PostID        *uint     `json:"post_id,omitempty" gorm:"column:post_id"`
}

// NotificationStats represents notification statistics
type NotificationStats struct {
	TotalCount    int64 `json:"total_count"`
	UnreadCount   int64 `json:"unread_count"`
	AppCount      int64 `json:"app_count"`
	PostCount     int64 `json:"post_count"`
	ScheduleCount int64 `json:"schedule_count"`
}
