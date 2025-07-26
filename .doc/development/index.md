# Development Guide

Przewodnik dla deweloperów pracujących z kodem SocGo - od setupu środowiska po deployment.

## 🛠️ Środowisko deweloperskie

### Wymagania
- **Go**: 1.24.3 lub nowszy
- **Git**: Do kontroli wersji
- **Node.js**: 18+ (do dokumentacji VitePress)
- **SQLite**: Wbudowane w Go
- **Make**: Opcjonalnie dla build scripts

### Setup developmentu

```bash
# 1. Clone repository
git clone https://github.com/tkowalski/socgo.git
cd socgo

# 2. Install dependencies
go mod download
go mod verify

# 3. Setup config
cp config.yml.example config.yml
# Edit config.yml with your API keys

# 4. Run tests
go test ./...

# 5. Start development server
go run main.go

# 6. Setup documentation (optional)
cd .doc
npm install
npm run docs:dev
```

## 🏗️ Architektura kodu

### Struktura katalogów

```
socgo/
├── main.go                     # Entry point
├── config.yml                  # Configuration
├── Makefile                    # Build scripts
├── Dockerfile                  # Container config
├── .gitignore                  # Git ignore rules
├── go.mod                      # Go modules
├── go.sum                      # Dependencies checksum
├── internal/                   # Application code (private)
│   ├── data/                  # Data layer
│   │   ├── config/           # Configuration structs
│   │   ├── database/         # Database models
│   │   ├── oauth/            # OAuth structs
│   │   └── provider/         # Provider definitions
│   ├── handlers/             # HTTP handlers (controllers)
│   │   ├── internal/         # Handler utilities
│   │   ├── calendar.go       # Calendar endpoints
│   │   ├── oauth.go          # OAuth flow
│   │   ├── post.go           # Post management
│   │   ├── setting.go        # Settings page
│   │   └── ...
│   ├── service/              # Business logic layer
│   │   ├── database/         # Database management
│   │   ├── dependency/       # Dependency injection
│   │   ├── notifications/    # Notification service
│   │   ├── oauth/            # OAuth service
│   │   ├── post/             # Post publishing service
│   │   ├── scheduler/        # Job scheduler
│   │   └── server/           # HTTP server
│   └── social/               # Social media integrations
│       ├── facebook/         # Facebook provider
│       ├── instagram/        # Instagram provider
│       └── tiktok/           # TikTok provider
├── web/                       # Frontend assets
│   ├── static/               # Static files (CSS, JS)
│   └── templates/            # Templ templates
├── uploads/                   # User uploaded files
└── .doc/                     # Documentation (VitePress)
```

### Wzorce projektowe

#### 1. Dependency Injection
```go
// internal/service/dependency/container.go
type Container struct {
    dbManager       *database.Manager
    oauthService    *oauth.Service
    postService     *post.ProviderService
    scheduler       *scheduler.Scheduler
}

func (c *Container) GetPostHandler() *handlers.PostHandler {
    return handlers.NewPostHandler(c.dbManager, c.postService)
}
```

#### 2. Repository Pattern
```go
// Abstrakcja dostępu do danych
type PostRepository interface {
    Create(post *Post) error
    GetByID(id uint) (*Post, error)
    GetByUserID(userID string) ([]*Post, error)
    Update(post *Post) error
    Delete(id uint) error
}
```

#### 3. Service Layer
```go
// Enkapsulacja logiki biznesowej
type PostService struct {
    repo     PostRepository
    provider ProviderService
    notifier NotificationService
}

func (s *PostService) PublishPost(userID string, content string) error {
    // Business logic here
}
```

## 🔄 Workflow deweloperski

### Branch Strategy

```
main                    # Production-ready code
├── develop            # Integration branch
├── feature/oauth-v2   # Feature branches
├── bugfix/rate-limit  # Bug fixes
└── hotfix/security    # Critical fixes
```

### Commit Convention

```bash
# Format: type(scope): description
feat(oauth): add TikTok provider support
fix(database): resolve connection pool leak
docs(api): update OAuth endpoint documentation
test(post): add integration tests for publishing
refactor(handlers): extract common validation logic
```

### Code Review Process

1. **Feature development** w feature branch
2. **Self-review** kodu i testów
3. **Create Pull Request** z opisem zmian
4. **Code review** od innych developerów
5. **Address feedback** i update PR
6. **Merge** po akceptacji

## 🧪 Testing

### Test Structure

```
internal/
├── handlers/
│   ├── post_test.go           # Unit tests
│   └── integration_test.go    # Integration tests
├── service/
│   ├── oauth/
│   │   ├── service_test.go
│   │   └── mock_test.go
│   └── post/
│       ├── service_test.go
│       └── provider_test.go
└── social/
    ├── facebook/
    │   ├── oauth_test.go
    │   └── post_test.go
    └── ...
```

### Rodzaje testów

#### Unit Tests
```go
func TestPostHandler_Create(t *testing.T) {
    // Setup
    mockDB := &MockDatabase{}
    mockProvider := &MockProvider{}
    handler := NewPostHandler(mockDB, mockProvider)
    
    // Test
    req := httptest.NewRequest("POST", "/posts", strings.NewReader(jsonPayload))
    resp := httptest.NewRecorder()
    handler.HandlePost(resp, req)
    
    // Assert
    assert.Equal(t, 201, resp.Code)
    assert.Contains(t, resp.Body.String(), "success")
}
```

#### Integration Tests
```go
func TestPostPublishing_Integration(t *testing.T) {
    // Setup real database and mocked external APIs
    db := setupTestDatabase(t)
    defer cleanupTestDatabase(t, db)
    
    // Test full flow
    post := createTestPost()
    err := publishPost(post)
    
    assert.NoError(t, err)
    // Verify database state
    // Verify external API calls
}
```

#### E2E Tests
```go
func TestFullWorkflow_E2E(t *testing.T) {
    // Start test server
    server := startTestServer(t)
    defer server.Close()
    
    // Test OAuth flow
    authURL := getAuthURL(server.URL)
    // ... simulate OAuth callback
    
    // Test post creation
    resp := createPost(server.URL, authToken, postData)
    assert.Equal(t, 201, resp.StatusCode)
}
```

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./internal/handlers

# With coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Integration tests only
go test -tags=integration ./...

# Verbose output
go test -v ./...

# Parallel execution
go test -p 4 ./...
```

## 🔧 Build i Deployment

### Makefile Commands

```makefile
# Development
.PHONY: dev
dev:
	go run main.go

.PHONY: test
test:
	go test ./...

.PHONY: test-coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Build
.PHONY: build
build:
	go build -o socgo main.go

.PHONY: build-all
build-all:
	GOOS=linux GOARCH=amd64 go build -o dist/socgo-linux-amd64 main.go
	GOOS=darwin GOARCH=amd64 go build -o dist/socgo-darwin-amd64 main.go
	GOOS=windows GOARCH=amd64 go build -o dist/socgo-windows-amd64.exe main.go

# Docker
.PHONY: docker-build
docker-build:
	docker build -t socgo:latest .

.PHONY: docker-run
docker-run:
	docker run -p 8080:8080 -v $(PWD)/config.yml:/app/config.yml socgo:latest

# Cleanup
.PHONY: clean
clean:
	rm -f socgo
	rm -rf dist/
	rm -f coverage.out
```

### Environment Variables

```bash
# Development
export SOCGO_ENV=development
export SOCGO_LOG_LEVEL=debug

# Production
export SOCGO_ENV=production
export SOCGO_LOG_LEVEL=info
export SOCGO_CONFIG_PATH=/etc/socgo/config.yml
```

### Docker Multi-stage Build

```dockerfile
# Build stage
FROM golang:1.24.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o socgo main.go

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/socgo .
COPY --from=builder /app/web ./web
CMD ["./socgo"]
```

## 📊 Monitoring i Debugging

### Logging

```go
// Structured logging
import "log/slog"

func (h *PostHandler) HandlePost(w http.ResponseWriter, r *http.Request) {
    logger := slog.With(
        "handler", "PostHandler",
        "method", r.Method,
        "path", r.URL.Path,
    )
    
    logger.Info("Processing post request")
    
    // ... handle request
    
    if err != nil {
        logger.Error("Post creation failed", 
            "error", err,
            "user_id", userID,
        )
        http.Error(w, "Internal server error", 500)
        return
    }
    
    logger.Info("Post created successfully", "post_id", post.ID)
}
```

### Performance Monitoring

```go
// Request timing middleware
func TimingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        next.ServeHTTP(w, r)
        
        duration := time.Since(start)
        slog.Info("Request completed",
            "method", r.Method,
            "path", r.URL.Path,
            "duration", duration,
        )
    })
}
```

### Health Checks

```go
func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
    health := map[string]interface{}{
        "status": "ok",
        "timestamp": time.Now(),
        "version": version,
        "services": map[string]string{
            "database": s.checkDatabase(),
            "oauth": s.checkOAuth(),
            "providers": s.checkProviders(),
        },
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(health)
}
```

## 🔐 Security

### Input Validation

```go
func validatePostRequest(req *PostRequest) error {
    if req.ProviderID == 0 {
        return errors.New("provider_id is required")
    }
    
    if len(req.Content) == 0 {
        return errors.New("content is required")
    }
    
    if len(req.Content) > 10000 {
        return errors.New("content too long")
    }
    
    // Sanitize HTML
    req.Content = html.EscapeString(req.Content)
    
    return nil
}
```

### OAuth Token Security

```go
// Encrypt tokens before storage
func encryptToken(token string, key []byte) (string, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return "", err
    }
    
    // Use AES-GCM for authenticated encryption
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }
    
    ciphertext := gcm.Seal(nonce, nonce, []byte(token), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}
```

## 📈 Performance Guidelines

### Database Optimization

```go
// Use database indexes
type Post struct {
    ID         uint      `gorm:"primaryKey"`
    UserID     string    `gorm:"index"`        // Index for user queries
    ProviderID uint      `gorm:"index"`        // Index for provider queries
    CreatedAt  time.Time `gorm:"index"`        // Index for date queries
    Content    string    `gorm:"not null"`
}

// Efficient queries with preloading
func GetPostsWithProvider(db *gorm.DB, userID string) ([]*Post, error) {
    var posts []*Post
    err := db.Preload("Provider").
        Where("user_id = ?", userID).
        Order("created_at DESC").
        Limit(50).
        Find(&posts).Error
    return posts, err
}
```

### Memory Management

```go
// Use sync.Pool for frequent allocations
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 1024)
    },
}

func processLargeFile(file io.Reader) error {
    buffer := bufferPool.Get().([]byte)
    defer bufferPool.Put(buffer)
    
    // Process file with pooled buffer
    return nil
}
```

### Caching Strategy

```go
// Simple in-memory cache for provider configs
type ProviderCache struct {
    mu    sync.RWMutex
    cache map[string]*Provider
    ttl   time.Duration
}

func (c *ProviderCache) Get(key string) (*Provider, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    provider, exists := c.cache[key]
    return provider, exists
}
```

## 🚀 Contributing Guidelines

### Code Style

```go
// Use gofmt and golint
go fmt ./...
golint ./...

// Follow Go conventions
// - Package names: lowercase, single word
// - Function names: CamelCase, start with capital if exported
// - Variable names: camelCase
// - Constants: UPPER_CASE or CamelCase
```

### Documentation

```go
// Document all exported functions
// Package post provides social media post publishing functionality.
package post

// PublishPost publishes content to a social media provider.
// It returns the external post ID and any error that occurred.
//
// The content is validated before publishing and may be modified
// to comply with provider requirements.
func PublishPost(provider Provider, content string) (string, error) {
    // Implementation
}
```

### Error Handling

```go
// Use typed errors for better error handling
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

// Wrap errors for context
func processPost(post *Post) error {
    if err := validatePost(post); err != nil {
        return fmt.Errorf("post validation failed: %w", err)
    }
    
    if err := publishPost(post); err != nil {
        return fmt.Errorf("post publishing failed: %w", err)
    }
    
    return nil
}
```

### Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests added/updated
```