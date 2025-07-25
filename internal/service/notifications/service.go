package notifications

import (
	"fmt"
	"time"

	data_database "github.com/tkowalski/socgo/internal/data/database"
	"github.com/tkowalski/socgo/internal/service/database"
)

type Service struct {
	dbManager *database.Manager
}

func NewService(dbManager *database.Manager) *Service {
	return &Service{
		dbManager: dbManager,
	}
}

// CreateNotification creates a new notification
func (s *Service) CreateNotification(userID, notificationType, category, title, message string, postID *uint) error {
	db, err := s.dbManager.GetDB(userID)
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	// Generate group ID based on type and post ID
	groupID := s.generateGroupID(notificationType, postID)

	// Check if group exists and is active
	var existingGroup data_database.Notification
	if err := db.Where("user_id = ? AND group_id = ? AND is_active = ?", userID, groupID, true).First(&existingGroup).Error; err == nil {
		// Group exists and is active, reactivate all notifications in the group
		if err := db.Model(&data_database.Notification{}).
			Where("user_id = ? AND group_id = ?", userID, groupID).
			Updates(map[string]interface{}{
				"is_read":   false,
				"is_active": true,
			}).Error; err != nil {
			return fmt.Errorf("failed to reactivate group: %w", err)
		}
	}

	notification := data_database.Notification{
		UserID:    userID,
		Type:      notificationType,
		Category:  category,
		Title:     title,
		Message:   message,
		PostID:    postID,
		GroupID:   groupID,
		IsRead:    false,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return db.Create(&notification).Error
}

// GetNotificationGroups returns grouped notifications for a user
func (s *Service) GetNotificationGroups(userID, notificationType string) ([]data_database.NotificationGroup, error) {
	db, err := s.dbManager.GetDB(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	// First, try to get raw results to debug
	var rawResults []map[string]interface{}

	baseQuery := `
		SELECT 
			group_id,
			type,
			title,
			message as latest_message,
			COUNT(*) as count,
			SUM(CASE WHEN is_read = 0 THEN 1 ELSE 0 END) as unread_count,
			MAX(is_active) as is_active,
			MAX(created_at) as latest_at,
			post_id
		FROM notifications 
		WHERE user_id = ? AND is_active = 1
	`

	if notificationType != "" {
		baseQuery += " AND type = ?"
		baseQuery += " GROUP BY group_id ORDER BY latest_at DESC"
		err = db.Raw(baseQuery, userID, notificationType).Scan(&rawResults).Error
	} else {
		baseQuery += " GROUP BY group_id ORDER BY latest_at DESC"
		err = db.Raw(baseQuery, userID).Scan(&rawResults).Error
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get notification groups: %w", err)
	}

	// Debug: print raw results (commented out for production)
	// fmt.Printf("Found %d raw notification groups for user %s\n", len(rawResults), userID)
	// for i, result := range rawResults {
	// 	fmt.Printf("Raw result %d: %+v\n", i, result)
	// }

	// Convert raw results to NotificationGroup structs
	var groups []data_database.NotificationGroup
	for _, raw := range rawResults {
		// Helper function to safely extract values
		getString := func(key string) string {
			if val, ok := raw[key]; ok && val != nil {
				if str, ok := val.(string); ok {
					return str
				}
			}
			return ""
		}

		getInt := func(key string) int {
			if val, ok := raw[key]; ok && val != nil {
				if ptr, ok := val.(*interface{}); ok && ptr != nil {
					if intVal, ok := (*ptr).(int64); ok {
						return int(intVal)
					}
				}
			}
			return 0
		}

		getBool := func(key string) bool {
			if val, ok := raw[key]; ok && val != nil {
				if ptr, ok := val.(*interface{}); ok && ptr != nil {
					if intVal, ok := (*ptr).(int64); ok {
						return intVal == 1
					}
				}
			}
			return false
		}

		getTime := func(key string) time.Time {
			if val, ok := raw[key]; ok && val != nil {
				if ptr, ok := val.(*interface{}); ok && ptr != nil {
					if timeVal, ok := (*ptr).(time.Time); ok {
						return timeVal
					}
				}
			}
			return time.Now()
		}

		group := data_database.NotificationGroup{
			GroupID:       getString("group_id"),
			Type:          getString("type"),
			Title:         getString("title"),
			LatestMessage: getString("latest_message"),
			Count:         getInt("count"),
			UnreadCount:   getInt("unread_count"),
			IsActive:      getBool("is_active"),
			LatestAt:      getTime("latest_at"),
			PostID:        nil, // Handle NULL post_id
		}

		groups = append(groups, group)
	}

	return groups, nil
}

// GetNotificationsByGroup returns all notifications in a group
func (s *Service) GetNotificationsByGroup(userID, groupID string) ([]data_database.Notification, error) {
	db, err := s.dbManager.GetDB(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	var notifications []data_database.Notification
	err = db.Where("user_id = ? AND group_id = ? AND is_active = ?", userID, groupID, true).
		Order("created_at DESC").
		Find(&notifications).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get notifications by group: %w", err)
	}

	return notifications, nil
}

// MarkGroupAsRead marks all notifications in a group as read
func (s *Service) MarkGroupAsRead(userID, groupID string) error {
	db, err := s.dbManager.GetDB(userID)
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	return db.Model(&data_database.Notification{}).
		Where("user_id = ? AND group_id = ?", userID, groupID).
		Update("is_read", true).Error
}

// GetNotificationStats returns notification statistics for a user
func (s *Service) GetNotificationStats(userID string) (*data_database.NotificationStats, error) {
	db, err := s.dbManager.GetDB(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	var stats data_database.NotificationStats

	// Total count
	if err := db.Model(&data_database.Notification{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Count(&stats.TotalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Unread count
	if err := db.Model(&data_database.Notification{}).
		Where("user_id = ? AND is_active = ? AND is_read = ?", userID, true, false).
		Count(&stats.UnreadCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get unread count: %w", err)
	}

	// Type counts
	if err := db.Model(&data_database.Notification{}).
		Where("user_id = ? AND is_active = ? AND type = ?", userID, true, "app").
		Count(&stats.AppCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get app count: %w", err)
	}

	if err := db.Model(&data_database.Notification{}).
		Where("user_id = ? AND is_active = ? AND type = ?", userID, true, "post").
		Count(&stats.PostCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get post count: %w", err)
	}

	if err := db.Model(&data_database.Notification{}).
		Where("user_id = ? AND is_active = ? AND type = ?", userID, true, "schedule").
		Count(&stats.ScheduleCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get schedule count: %w", err)
	}

	return &stats, nil
}

// generateGroupID generates a unique group ID based on type and post ID
func (s *Service) generateGroupID(notificationType string, postID *uint) string {
	if postID != nil {
		return fmt.Sprintf("%s_post_%d", notificationType, *postID)
	}
	return fmt.Sprintf("%s_%d", notificationType, time.Now().Unix())
}
