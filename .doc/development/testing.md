# Testing Guide

Kompletny przewodnik testowania w SocGo - od unit testów po testy end-to-end.

## 🧪 Strategia testowa

### Piramida testów

```
                /\
               /  \
              / E2E \     ← Mało, powolne, wysokopoziomowe
             /______\
            /        \
           / Integration\ ← Średnio, średnie, middle-level
          /______________\
         /                \
        /   Unit Tests     \ ← Dużo, szybkie, niskopoziomowe
       /____________________\
```

### Test Coverage Targets
- **Unit Tests**: 85%+ coverage
- **Integration Tests**: Krytyczne ścieżki
- **E2E Tests**: Główne user journeys

## 🔧 Setup testowego

### Test Dependencies

```go
// go.mod test dependencies
require (
    github.com/stretchr/testify v1.8.4
    github.com/golang/mock v1.6.0
    github.com/testcontainers/testcontainers-go v0.26.0
    github.com/gavv/httpexpect/v2 v2.16.0
)
```

### Test Configuration

```go
// internal/testing/config.go
type TestConfig struct {
    Database struct {
        URL      string `yaml:"url"`
        InMemory bool   `yaml:"in_memory"`
    } `yaml:"database"`
    
    OAuth struct {
        MockMode bool              `yaml:"mock_mode"`
        Providers map[string]struct {
            ClientID     string `yaml:"client_id"`
            ClientSecret string `yaml:"client_secret"`
        } `yaml:"providers"`
    } `yaml:"oauth"`
}

func LoadTestConfig() (*TestConfig, error) {
    config := &TestConfig{}
    data, err := os.ReadFile("testdata/config.test.yml")
    if err != nil {
        return nil, err
    }
    return config, yaml.Unmarshal(data, config)
}
```

## 📝 Unit Tests

### Test Structure

```go
// internal/service/post/service_test.go
package post

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"
)

type PostServiceTestSuite struct {
    suite.Suite
    service *Service
    mockRepo *MockRepository
    mockProvider *MockProvider
}

func (suite *PostServiceTestSuite) SetupTest() {
    suite.mockRepo = NewMockRepository()
    suite.mockProvider = NewMockProvider()
    suite.service = NewService(suite.mockRepo, suite.mockProvider)
}

func (suite *PostServiceTestSuite) TestCreatePost_Success() {
    // Arrange
    content := "Test post content"
    userID := "user123"
    expectedPost := &Post{
        ID: 1,
        Content: content,
        UserID: userID,
    }
    
    suite.mockRepo.On("Create", mock.AnythingOfType("*Post")).
        Return(nil).
        Run(func(args mock.Arguments) {
            post := args.Get(0).(*Post)
            post.ID = 1
        })
    
    // Act
    result, err := suite.service.CreatePost(content, userID)
    
    // Assert
    assert.NoError(suite.T(), err)
    assert.Equal(suite.T(), uint(1), result.ID)
    assert.Equal(suite.T(), content, result.Content)
    suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *PostServiceTestSuite) TestCreatePost_ValidationError() {
    // Arrange
    content := "" // Invalid empty content
    userID := "user123"
    
    // Act
    result, err := suite.service.CreatePost(content, userID)
    
    // Assert
    assert.Error(suite.T(), err)
    assert.Nil(suite.T(), result)
    assert.Contains(suite.T(), err.Error(), "content is required")
}

func TestPostServiceTestSuite(t *testing.T) {
    suite.Run(t, new(PostServiceTestSuite))
}
```

### Mock Generation

```go
//go:generate mockgen -source=repository.go -destination=mocks/mock_repository.go

// Repository interface for mocking
type Repository interface {
    Create(post *Post) error
    GetByID(id uint) (*Post, error)
    GetByUserID(userID string) ([]*Post, error)
    Update(post *Post) error
    Delete(id uint) error
}

// Provider interface for mocking
type Provider interface {
    PublishPost(content string, media []Media) (string, error)
    ValidateCredentials() error
    GetUserInfo() (*UserInfo, error)
}
```

### Table-driven Tests

```go
func TestValidatePost(t *testing.T) {
    tests := []struct {
        name        string
        post        *Post
        expectedErr string
    }{
        {
            name: "valid post",
            post: &Post{
                Content: "Valid content",
                UserID:  "user123",
            },
            expectedErr: "",
        },
        {
            name: "empty content",
            post: &Post{
                Content: "",
                UserID:  "user123",
            },
            expectedErr: "content is required",
        },
        {
            name: "content too long",
            post: &Post{
                Content: strings.Repeat("x", 10001),
                UserID:  "user123",
            },
            expectedErr: "content too long",
        },
        {
            name: "empty user ID",
            post: &Post{
                Content: "Valid content",
                UserID:  "",
            },
            expectedErr: "user ID is required",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidatePost(tt.post)
            
            if tt.expectedErr == "" {
                assert.NoError(t, err)
            } else {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedErr)
            }
        })
    }
}
```

## 🔗 Integration Tests

### Database Integration Tests

```go
//go:build integration
// +build integration

package database

import (
    "testing"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/sqlite"
)

type DatabaseIntegrationTestSuite struct {
    suite.Suite
    container testcontainers.Container
    db        *gorm.DB
    manager   *Manager
}

func (suite *DatabaseIntegrationTestSuite) SetupSuite() {
    ctx := context.Background()
    
    // Start SQLite test container
    container, err := sqlite.RunContainer(ctx,
        testcontainers.WithImage("sqlite:latest"),
    )
    suite.Require().NoError(err)
    suite.container = container
    
    // Get connection string
    connStr, err := container.ConnectionString(ctx)
    suite.Require().NoError(err)
    
    // Setup database manager
    suite.manager = NewManager(connStr)
    suite.db, err = suite.manager.GetDB("test_user")
    suite.Require().NoError(err)
}

func (suite *DatabaseIntegrationTestSuite) TearDownSuite() {
    if suite.container != nil {
        suite.container.Terminate(context.Background())
    }
}

func (suite *DatabaseIntegrationTestSuite) SetupTest() {
    // Clean database before each test
    suite.db.Exec("DELETE FROM posts")
    suite.db.Exec("DELETE FROM providers")
}

func (suite *DatabaseIntegrationTestSuite) TestPostCRUD() {
    // Create
    post := &Post{
        Content: "Test content",
        UserID:  "test_user",
        ProviderID: 1,
    }
    
    err := suite.db.Create(post).Error
    suite.NoError(err)
    suite.NotZero(post.ID)
    
    // Read
    var foundPost Post
    err = suite.db.First(&foundPost, post.ID).Error
    suite.NoError(err)
    suite.Equal(post.Content, foundPost.Content)
    
    // Update
    foundPost.Content = "Updated content"
    err = suite.db.Save(&foundPost).Error
    suite.NoError(err)
    
    // Verify update
    var updatedPost Post
    err = suite.db.First(&updatedPost, post.ID).Error
    suite.NoError(err)
    suite.Equal("Updated content", updatedPost.Content)
    
    // Delete
    err = suite.db.Delete(&foundPost).Error
    suite.NoError(err)
    
    // Verify deletion
    var deletedPost Post
    err = suite.db.First(&deletedPost, post.ID).Error
    suite.Error(err) // Should not find deleted post
}

func TestDatabaseIntegrationTestSuite(t *testing.T) {
    suite.Run(t, new(DatabaseIntegrationTestSuite))
}
```

### HTTP Handler Integration Tests

```go
//go:build integration
// +build integration

package handlers

func TestPostHandler_Integration(t *testing.T) {
    // Setup test server
    server := setupTestServer(t)
    defer server.Close()
    
    // Setup test client
    e := httpexpect.Default(t, server.URL)
    
    t.Run("Create Post Success", func(t *testing.T) {
        postData := map[string]interface{}{
            "provider_id": 1,
            "content":     "Test post content",
            "schedule_at": "now",
        }
        
        resp := e.POST("/posts").
            WithJSON(postData).
            Expect().
            Status(http.StatusCreated).
            JSON()
        
        resp.Object().
            HasValue("status", "published").
            HasValue("content", "Test post content").
            HasKey("id")
    })
    
    t.Run("Create Post Validation Error", func(t *testing.T) {
        postData := map[string]interface{}{
            "provider_id": 0, // Invalid
            "content":     "",
        }
        
        e.POST("/posts").
            WithJSON(postData).
            Expect().
            Status(http.StatusBadRequest).
            Body().Contains("provider_id is required")
    })
}

func setupTestServer(t *testing.T) *httptest.Server {
    // Setup test database
    db := setupTestDB(t)
    
    // Setup test container
    container := dependency.NewContainer()
    container.SetDB(db)
    
    // Setup handlers
    postHandler := container.GetPostHandler()
    
    // Setup router
    router := mux.NewRouter()
    router.HandleFunc("/posts", postHandler.HandlePost).Methods("POST")
    
    return httptest.NewServer(router)
}
```

### OAuth Integration Tests

```go
//go:build integration
// +build integration

package oauth

func TestOAuthFlow_Integration(t *testing.T) {
    tests := []struct {
        name     string
        provider string
        mockMode bool
    }{
        {"Facebook OAuth", "facebook", true},
        {"Instagram OAuth", "instagram", true},
        {"TikTok OAuth", "tiktok", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mock OAuth server if needed
            var mockServer *httptest.Server
            if tt.mockMode {
                mockServer = setupMockOAuthServer(t, tt.provider)
                defer mockServer.Close()
            }
            
            // Test OAuth flow
            service := setupOAuthService(t, tt.provider, mockServer)
            
            // Test get auth URL
            authURL, err := service.GetAuthURL("test_state")
            assert.NoError(t, err)
            assert.Contains(t, authURL, "oauth")
            
            // Test token exchange (mocked)
            if tt.mockMode {
                token, err := service.ExchangeCode("test_code")
                assert.NoError(t, err)
                assert.NotEmpty(t, token.AccessToken)
            }
        })
    }
}

func setupMockOAuthServer(t *testing.T, provider string) *httptest.Server {
    mux := http.NewServeMux()
    
    // Mock token endpoint
    mux.HandleFunc("/oauth/access_token", func(w http.ResponseWriter, r *http.Request) {
        response := map[string]interface{}{
            "access_token": "mock_access_token",
            "token_type":   "Bearer",
            "expires_in":   3600,
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
    })
    
    return httptest.NewServer(mux)
}
```

## 🌐 End-to-End Tests

### Browser Tests with Selenium

```go
//go:build e2e
// +build e2e

package e2e

import (
    "testing"
    "github.com/tebeka/selenium"
    "github.com/tebeka/selenium/chrome"
)

type E2ETestSuite struct {
    suite.Suite
    driver   selenium.WebDriver
    service  *selenium.Service
    baseURL  string
}

func (suite *E2ETestSuite) SetupSuite() {
    // Start Selenium service
    opts := []selenium.ServiceOption{}
    service, err := selenium.NewChromeDriverService("./chromedriver", 9515, opts...)
    suite.Require().NoError(err)
    suite.service = service
    
    // Setup Chrome driver
    caps := selenium.Capabilities{"browserName": "chrome"}
    chromeCaps := chrome.Capabilities{
        Args: []string{
            "--headless",
            "--no-sandbox",
            "--disable-dev-shm-usage",
        },
    }
    caps.AddChrome(chromeCaps)
    
    driver, err := selenium.NewRemote(caps, "http://localhost:9515")
    suite.Require().NoError(err)
    suite.driver = driver
    
    // Start test server
    suite.baseURL = startTestServer()
}

func (suite *E2ETestSuite) TearDownSuite() {
    if suite.driver != nil {
        suite.driver.Quit()
    }
    if suite.service != nil {
        suite.service.Stop()
    }
}

func (suite *E2ETestSuite) TestCompletePostWorkflow() {
    // Navigate to home page
    err := suite.driver.Get(suite.baseURL)
    suite.NoError(err)
    
    // Verify page loaded
    title, err := suite.driver.Title()
    suite.NoError(err)
    suite.Contains(title, "SocGo")
    
    // Fill post content
    contentField, err := suite.driver.FindElement(selenium.ByID, "post-content")
    suite.NoError(err)
    
    err = contentField.SendKeys("Test post from E2E test")
    suite.NoError(err)
    
    // Select provider
    providerCheckbox, err := suite.driver.FindElement(selenium.ByXPATH, "//input[@type='checkbox'][@value='1']")
    suite.NoError(err)
    
    err = providerCheckbox.Click()
    suite.NoError(err)
    
    // Submit post
    submitButton, err := suite.driver.FindElement(selenium.ByID, "publish-button")
    suite.NoError(err)
    
    err = submitButton.Click()
    suite.NoError(err)
    
    // Wait for success message
    suite.Eventually(func() bool {
        successMessage, err := suite.driver.FindElement(selenium.ByClass, "success-message")
        if err != nil {
            return false
        }
        
        text, err := successMessage.Text()
        return err == nil && strings.Contains(text, "Post published successfully")
    }, 10*time.Second, 500*time.Millisecond, "Success message should appear")
}

func TestE2ETestSuite(t *testing.T) {
    suite.Run(t, new(E2ETestSuite))
}
```

### API E2E Tests

```go
//go:build e2e
// +build e2e

func TestPostAPI_E2E(t *testing.T) {
    baseURL := getTestServerURL()
    client := httpexpect.Default(t, baseURL)
    
    t.Run("Complete Post Lifecycle", func(t *testing.T) {
        // 1. Create post
        postResp := client.POST("/posts").
            WithJSON(map[string]interface{}{
                "provider_id": 1,
                "content":     "E2E test post",
                "schedule_at": "now",
            }).
            Expect().
            Status(http.StatusCreated).
            JSON()
        
        postID := postResp.Object().Value("id").Number().Raw()
        
        // 2. Verify post in history
        historyResp := client.GET("/posts/history").
            Expect().
            Status(http.StatusOK).
            JSON()
        
        posts := historyResp.Object().Value("posts").Array()
        posts.Length().Gt(0)
        
        // Find our post
        found := false
        for _, item := range posts.Iter() {
            if item.Object().Value("id").Number().Raw() == postID {
                item.Object().Value("content").String().Equal("E2E test post")
                found = true
                break
            }
        }
        assert.True(t, found, "Post should be found in history")
        
        // 3. Get post details
        client.GET(fmt.Sprintf("/posts/details/%v", postID)).
            Expect().
            Status(http.StatusOK).
            Body().Contains("E2E test post")
        
        // 4. Delete post
        client.DELETE(fmt.Sprintf("/posts/%v", postID)).
            Expect().
            Status(http.StatusOK)
        
        // 5. Verify post deleted
        client.GET(fmt.Sprintf("/posts/details/%v", postID)).
            Expect().
            Status(http.StatusNotFound)
    })
}
```

## 📊 Performance Tests

### Benchmark Tests

```go
func BenchmarkPostService_CreatePost(b *testing.B) {
    service := setupBenchmarkService(b)
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        counter := 0
        for pb.Next() {
            content := fmt.Sprintf("Benchmark post %d", counter)
            userID := fmt.Sprintf("user%d", counter%100)
            
            _, err := service.CreatePost(content, userID)
            if err != nil {
                b.Fatal(err)
            }
            counter++
        }
    })
}

func BenchmarkDatabase_Insert(b *testing.B) {
    db := setupBenchmarkDB(b)
    
    posts := make([]*Post, b.N)
    for i := 0; i < b.N; i++ {
        posts[i] = &Post{
            Content: fmt.Sprintf("Benchmark post %d", i),
            UserID:  fmt.Sprintf("user%d", i%100),
        }
    }
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        i := 0
        for pb.Next() {
            err := db.Create(posts[i]).Error
            if err != nil {
                b.Fatal(err)
            }
            i++
        }
    })
}
```

### Load Tests

```go
//go:build load
// +build load

func TestAPI_LoadTest(t *testing.T) {
    baseURL := getTestServerURL()
    
    // Configuration
    const (
        numUsers = 100
        duration = 30 * time.Second
        rampUp   = 5 * time.Second
    )
    
    var wg sync.WaitGroup
    results := make(chan TestResult, numUsers*10)
    
    // Ramp up users
    userDelay := rampUp / time.Duration(numUsers)
    
    for i := 0; i < numUsers; i++ {
        wg.Add(1)
        go func(userID int) {
            defer wg.Done()
            
            client := httpexpect.Default(t, baseURL)
            startTime := time.Now()
            
            for time.Since(startTime) < duration {
                // Simulate user behavior
                testStart := time.Now()
                
                resp := client.POST("/posts").
                    WithJSON(map[string]interface{}{
                        "provider_id": 1,
                        "content":     fmt.Sprintf("Load test post from user %d", userID),
                        "schedule_at": "now",
                    }).
                    Expect()
                
                result := TestResult{
                    UserID:       userID,
                    Duration:     time.Since(testStart),
                    StatusCode:   resp.Raw().StatusCode,
                    Success:      resp.Raw().StatusCode == 201,
                }
                
                results <- result
                
                // Wait before next request
                time.Sleep(time.Millisecond * 100)
            }
        }(i)
        
        time.Sleep(userDelay)
    }
    
    // Wait for all users to finish
    go func() {
        wg.Wait()
        close(results)
    }()
    
    // Collect results
    var (
        totalRequests int
        successCount  int
        totalDuration time.Duration
        maxDuration   time.Duration
    )
    
    for result := range results {
        totalRequests++
        if result.Success {
            successCount++
        }
        
        totalDuration += result.Duration
        if result.Duration > maxDuration {
            maxDuration = result.Duration
        }
    }
    
    // Report results
    successRate := float64(successCount) / float64(totalRequests) * 100
    avgDuration := totalDuration / time.Duration(totalRequests)
    
    t.Logf("Load Test Results:")
    t.Logf("- Total Requests: %d", totalRequests)
    t.Logf("- Success Rate: %.2f%%", successRate)
    t.Logf("- Average Duration: %v", avgDuration)
    t.Logf("- Max Duration: %v", maxDuration)
    
    // Assertions
    assert.Greater(t, successRate, 95.0, "Success rate should be > 95%")
    assert.Less(t, avgDuration, time.Second, "Average response time should be < 1s")
}

type TestResult struct {
    UserID     int
    Duration   time.Duration
    StatusCode int
    Success    bool
}
```

## 🎯 Test Data Management

### Test Fixtures

```go
// testdata/fixtures.go
package testdata

type Fixtures struct {
    Posts     []*Post     `json:"posts"`
    Providers []*Provider `json:"providers"`
    Users     []*User     `json:"users"`
}

func LoadFixtures() (*Fixtures, error) {
    data, err := os.ReadFile("testdata/fixtures.json")
    if err != nil {
        return nil, err
    }
    
    var fixtures Fixtures
    err = json.Unmarshal(data, &fixtures)
    return &fixtures, err
}

func (f *Fixtures) SeedDatabase(db *gorm.DB) error {
    // Clear existing data
    db.Exec("DELETE FROM posts")
    db.Exec("DELETE FROM providers")
    
    // Seed providers
    for _, provider := range f.Providers {
        if err := db.Create(provider).Error; err != nil {
            return err
        }
    }
    
    // Seed posts
    for _, post := range f.Posts {
        if err := db.Create(post).Error; err != nil {
            return err
        }
    }
    
    return nil
}
```

### Factory Pattern for Test Data

```go
// internal/testing/factory.go
package testing

type PostFactory struct {
    counter int
}

func NewPostFactory() *PostFactory {
    return &PostFactory{}
}

func (f *PostFactory) Create(opts ...PostOption) *Post {
    f.counter++
    
    post := &Post{
        Content:    fmt.Sprintf("Test post %d", f.counter),
        UserID:     "test_user",
        ProviderID: 1,
        CreatedAt:  time.Now(),
        UpdatedAt:  time.Now(),
    }
    
    for _, opt := range opts {
        opt(post)
    }
    
    return post
}

type PostOption func(*Post)

func WithContent(content string) PostOption {
    return func(p *Post) {
        p.Content = content
    }
}

func WithUserID(userID string) PostOption {
    return func(p *Post) {
        p.UserID = userID
    }
}

func WithProviderID(providerID uint) PostOption {
    return func(p *Post) {
        p.ProviderID = providerID
    }
}

// Usage in tests
func TestSomething(t *testing.T) {
    factory := NewPostFactory()
    
    post1 := factory.Create()
    post2 := factory.Create(WithContent("Custom content"), WithUserID("user123"))
}
```

## 🚀 Running Tests

### Local Development

```bash
# Run all unit tests
go test ./...

# Run tests with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test package
go test ./internal/handlers

# Run specific test
go test -run TestPostHandler_Create ./internal/handlers

# Run integration tests
go test -tags=integration ./...

# Run E2E tests
go test -tags=e2e ./...

# Run load tests
go test -tags=load ./...

# Run benchmarks
go test -bench=. ./...

# Run with race detection
go test -race ./...

# Verbose output
go test -v ./...

# Parallel execution
go test -parallel 4 ./...
```

### CI/CD Pipeline

```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install dependencies
        run: go mod download
      
      - name: Run unit tests
        run: go test -race -coverprofile=coverage.out ./...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out

  integration-tests:
    runs-on: ubuntu-latest
    services:
      sqlite:
        image: sqlite:latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run integration tests
        run: go test -tags=integration ./...

  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install Chrome
        uses: browser-actions/setup-chrome@latest
      
      - name: Run E2E tests
        run: go test -tags=e2e ./...
```

## 📈 Test Metrics i Reporting

### Coverage Reporting

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Coverage by package
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep -v "total:"
```

### Test Results Parsing

```go
// internal/testing/reporter.go
package testing

type TestResults struct {
    TotalTests  int           `json:"total_tests"`
    PassedTests int           `json:"passed_tests"`
    FailedTests int           `json:"failed_tests"`
    Duration    time.Duration `json:"duration"`
    Coverage    float64       `json:"coverage"`
    Packages    []PackageResult `json:"packages"`
}

type PackageResult struct {
    Name     string        `json:"name"`
    Tests    int          `json:"tests"`
    Passed   int          `json:"passed"`
    Failed   int          `json:"failed"`
    Duration time.Duration `json:"duration"`
    Coverage float64       `json:"coverage"`
}

func ParseTestOutput(output []byte) (*TestResults, error) {
    // Parse go test JSON output
    // Implementation details...
}
```

## 🔍 Test Debugging

### Debug Failed Tests

```bash
# Run specific failing test with verbose output
go test -v -run TestFailingTest ./package

# Run with additional logging
SOCGO_LOG_LEVEL=debug go test -v ./...

# Run with debugger
dlv test -- -test.run TestFailingTest
```

### Test Isolation

```go
func TestWithIsolation(t *testing.T) {
    // Ensure test isolation
    oldEnv := os.Getenv("SOCGO_ENV")
    defer func() {
        os.Setenv("SOCGO_ENV", oldEnv)
    }()
    
    // Set test environment
    os.Setenv("SOCGO_ENV", "test")
    
    // Your test code here
}
```

## 📋 Best Practices

### Test Naming
```go
// Function under test: GetUser
// Test cases:
func TestGetUser_ValidID_ReturnsUser(t *testing.T) {}
func TestGetUser_InvalidID_ReturnsError(t *testing.T) {}
func TestGetUser_DatabaseError_ReturnsError(t *testing.T) {}
```

### Test Organization
```go
func TestUserService(t *testing.T) {
    t.Run("Create", func(t *testing.T) {
        t.Run("ValidData", func(t *testing.T) {})
        t.Run("InvalidEmail", func(t *testing.T) {})
        t.Run("DuplicateEmail", func(t *testing.T) {})
    })
    
    t.Run("Update", func(t *testing.T) {
        t.Run("ValidData", func(t *testing.T) {})
        t.Run("UserNotFound", func(t *testing.T) {})
    })
}
```

### Assertion Guidelines
```go
// Good: Specific assertions
assert.Equal(t, "expected", actual)
assert.Len(t, users, 3)
assert.Contains(t, result, "success")

// Avoid: Generic assertions
assert.True(t, len(users) == 3) // Use assert.Len instead
assert.True(t, strings.Contains(result, "success")) // Use assert.Contains
```

To znacznie rozszerza dokumentację testową i pokrywa wszystkie aspekty testowania w SocGo! 🧪