package database

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Content      string         `json:"content" gorm:"not null"`
	Title        string         `json:"title"`
	UserID       string         `json:"user_id" gorm:"not null;index"`
	ProviderID   uint           `json:"provider_id" gorm:"index"`
	Provider     Provider       `json:"provider" gorm:"foreignKey:ProviderID"`
	Settings     string         `json:"settings" gorm:"type:text"`       // Ustawienia providera jako JSON
	ExternalID   string         `json:"external_id" gorm:"index"`        // ID posta na platformie zewnętrznej (np. Facebook post ID)
	ExternalURL  string         `json:"external_url"`                    // URL do posta na platformie zewnętrznej
	ScheduledAt  *time.Time     `json:"scheduled_at" gorm:"index"`       // Data planowanej wysyłki
	PublishedAt  *time.Time     `json:"published_at"`                    // Data publikacji na platformie zewnętrznej
	Status       string         `json:"status" gorm:"default:'pending'"` // Status posta: pending, published, failed
	ErrorMessage string         `json:"error_message"`                   // Komunikat błędu jeśli publikacja się nie powiodła
	Media        []Media        `json:"media" gorm:"foreignKey:PostID"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

type Media struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	PostID    uint           `json:"post_id" gorm:"index"`
	FileName  string         `json:"file_name" gorm:"not null"`
	FilePath  string         `json:"file_path" gorm:"not null"`
	FileType  string         `json:"file_type" gorm:"not null"` // image/video
	FileSize  int64          `json:"file_size"`
	MimeType  string         `json:"mime_type"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

type Provider struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"not null"`
	Type      string         `json:"type" gorm:"not null"`
	Config    string         `json:"config" gorm:"type:text"`
	UserID    string         `json:"user_id" gorm:"not null;index"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

type ScheduledJob struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	JobType     string     `json:"job_type" gorm:"not null"`
	PayloadData string     `json:"payload_data" gorm:"type:text"`
	UserID      string     `json:"user_id" gorm:"not null;index"`
	ProviderID  uint       `json:"provider_id" gorm:"index"`
	Provider    Provider   `json:"provider" gorm:"foreignKey:ProviderID"`
	ScheduledAt time.Time  `json:"scheduled_at" gorm:"not null;index"`
	ExecutedAt  *time.Time `json:"executed_at,omitempty"`
	Status      string     `json:"status" gorm:"default:'pending'"`
	ErrorMsg    string     `json:"error_msg"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type APIToken struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Hash      string         `json:"-" gorm:"not null;uniqueIndex;type:varchar(64)"`
	UserID    string         `json:"user_id" gorm:"not null;index"`
	CreatedAt time.Time      `json:"created_at"`
	LastUsed  *time.Time     `json:"last_used,omitempty"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

const (
	JobStatusPending   = "pending"
	JobStatusExecuting = "executing"
	JobStatusCompleted = "completed"
	JobStatusFailed    = "failed"
)

const (
	PostStatusPending   = "pending"
	PostStatusPublished = "published"
	PostStatusFailed    = "failed"
)
