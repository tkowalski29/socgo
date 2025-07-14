package database

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Content     string    `json:"content" gorm:"not null"`
	Title       string    `json:"title"`
	UserID      string    `json:"user_id" gorm:"not null;index"`
	ProviderID  uint      `json:"provider_id" gorm:"index"`
	Provider    Provider  `json:"provider" gorm:"foreignKey:ProviderID"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

type Provider struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null;uniqueIndex"`
	Type      string    `json:"type" gorm:"not null"`
	Config    string    `json:"config" gorm:"type:text"`
	UserID    string    `json:"user_id" gorm:"not null;index"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

type ScheduledJob struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	JobType     string    `json:"job_type" gorm:"not null"`
	PayloadData string    `json:"payload_data" gorm:"type:text"`
	UserID      string    `json:"user_id" gorm:"not null;index"`
	ProviderID  uint      `json:"provider_id" gorm:"index"`
	Provider    Provider  `json:"provider" gorm:"foreignKey:ProviderID"`
	ScheduledAt time.Time `json:"scheduled_at" gorm:"not null;index"`
	ExecutedAt  *time.Time `json:"executed_at,omitempty"`
	Status      string    `json:"status" gorm:"default:'pending'"`
	ErrorMsg    string    `json:"error_msg"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

const (
	JobStatusPending   = "pending"
	JobStatusExecuting = "executing"
	JobStatusCompleted = "completed"
	JobStatusFailed    = "failed"
)