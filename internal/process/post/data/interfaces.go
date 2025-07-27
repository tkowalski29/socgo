package data

import (
	"net/http"

	data_database "github.com/tkowalski/socgo/internal/data/database"
	provider_pkg "github.com/tkowalski/socgo/internal/data/provider"
	"github.com/tkowalski/socgo/internal/service/database"
)

// Task represents a single responsibility task in the post creation process
type Task interface {
	Execute(ctx *PostContext) error
}

// PostContext contains all the data needed for post creation process
type PostContext struct {
	Request      *http.Request
	UserID       string
	DB           *database.Manager
	Content      string
	ScheduleAt   string
	Media        []provider_pkg.Media
	Settings     map[string]string
	ProviderIDs  []uint
	PayloadJSON  []byte
	Errors       []error
	PendingPosts []data_database.Post // For scheduler process
}

// PostProcess orchestrates the post creation process
type PostProcess interface {
	Execute(ctx *PostContext) error
}

// PostProcessImpl implements PostProcess
type PostProcessImpl struct {
	tasks []Task
}
