package database

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	manager := NewManager("/tmp/test_data")

	if manager == nil {
		t.Fatal("NewManager returned nil")
		return // ensure no nil dereference warning
	}

	if manager.dataDir != "/tmp/test_data" {
		t.Errorf("Expected dataDir to be '/tmp/test_data', got %s", manager.dataDir)
	}
}

func TestGetDB_CreatesNewUserDB(t *testing.T) {
	tmpDir := "/tmp/test_socgo_" + time.Now().Format("20060102_150405")
	defer os.RemoveAll(tmpDir)

	manager := NewManager(tmpDir)
	userID := "test_user_123"

	db, err := manager.GetDB(userID)
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	if db == nil {
		t.Fatal("GetDB returned nil database")
	}

	expectedPath := filepath.Join(tmpDir, userID+".db")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Database file was not created at %s", expectedPath)
	}
}

func TestGetDB_ReturnsExistingDB(t *testing.T) {
	tmpDir := "/tmp/test_socgo_" + time.Now().Format("20060102_150405")
	defer os.RemoveAll(tmpDir)

	manager := NewManager(tmpDir)
	userID := "test_user_456"

	db1, err := manager.GetDB(userID)
	if err != nil {
		t.Fatalf("First GetDB call failed: %v", err)
	}

	db2, err := manager.GetDB(userID)
	if err != nil {
		t.Fatalf("Second GetDB call failed: %v", err)
	}

	if db1 != db2 {
		t.Error("GetDB should return the same database instance for the same user")
	}
}

func TestDatabaseTables_CreatedSuccessfully(t *testing.T) {
	tmpDir := "/tmp/test_socgo_" + time.Now().Format("20060102_150405")
	defer os.RemoveAll(tmpDir)

	manager := NewManager(tmpDir)
	userID := "test_user_789"

	db, err := manager.GetDB(userID)
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	tables := []string{"posts", "providers", "scheduled_jobs", "api_tokens"}

	for _, table := range tables {
		if !db.Migrator().HasTable(table) {
			t.Errorf("Table %s was not created", table)
		}
	}
}

func TestDatabaseOperations_InsertAndRetrieve(t *testing.T) {
	tmpDir := "/tmp/test_socgo_" + time.Now().Format("20060102_150405")
	defer os.RemoveAll(tmpDir)

	manager := NewManager(tmpDir)
	userID := "test_user_ops"

	db, err := manager.GetDB(userID)
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	provider := Provider{
		Name:     "Test Provider",
		Type:     "twitter",
		Config:   "{\"api_key\": \"test\"}",
		UserID:   userID,
		IsActive: true,
	}

	if err := db.Create(&provider).Error; err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	var retrievedProvider Provider
	if err := db.First(&retrievedProvider, provider.ID).Error; err != nil {
		t.Fatalf("Failed to retrieve provider: %v", err)
	}

	if retrievedProvider.Name != provider.Name {
		t.Errorf("Expected provider name %s, got %s", provider.Name, retrievedProvider.Name)
	}

	post := Post{
		Content:    "Test post content",
		Title:      "Test Post",
		UserID:     userID,
		ProviderID: provider.ID,
	}

	if err := db.Create(&post).Error; err != nil {
		t.Fatalf("Failed to create post: %v", err)
	}

	var retrievedPost Post
	if err := db.Preload("Provider").First(&retrievedPost, post.ID).Error; err != nil {
		t.Fatalf("Failed to retrieve post: %v", err)
	}

	if retrievedPost.Content != post.Content {
		t.Errorf("Expected post content %s, got %s", post.Content, retrievedPost.Content)
	}

	if retrievedPost.Provider.Name != provider.Name {
		t.Errorf("Expected provider name %s, got %s", provider.Name, retrievedPost.Provider.Name)
	}
}

func TestUserDBExists(t *testing.T) {
	tmpDir := "/tmp/test_socgo_" + time.Now().Format("20060102_150405")
	defer os.RemoveAll(tmpDir)

	manager := NewManager(tmpDir)
	userID := "test_user_exists"

	if manager.UserDBExists(userID) {
		t.Error("UserDBExists should return false for non-existent user")
	}

	_, err := manager.GetDB(userID)
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	if !manager.UserDBExists(userID) {
		t.Error("UserDBExists should return true after creating user DB")
	}
}

func TestCloseDB(t *testing.T) {
	tmpDir := "/tmp/test_socgo_" + time.Now().Format("20060102_150405")
	defer os.RemoveAll(tmpDir)

	manager := NewManager(tmpDir)
	userID := "test_user_close"

	_, err := manager.GetDB(userID)
	if err != nil {
		t.Fatalf("GetDB failed: %v", err)
	}

	if err := manager.CloseDB(userID); err != nil {
		t.Fatalf("CloseDB failed: %v", err)
	}

	if len(manager.dbs) != 0 {
		t.Error("Database should be removed from manager after closing")
	}
}
