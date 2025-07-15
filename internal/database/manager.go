package database

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Manager struct {
	dataDir string
	dbs     map[string]*gorm.DB
	mutex   sync.RWMutex
}

func NewManager(dataDir string) *Manager {
	return &Manager{
		dataDir: dataDir,
		dbs:     make(map[string]*gorm.DB),
	}
}

func (m *Manager) GetDB(userID string) (*gorm.DB, error) {
	m.mutex.RLock()
	if db, exists := m.dbs[userID]; exists {
		m.mutex.RUnlock()
		return db, nil
	}
	m.mutex.RUnlock()

	return m.createOrOpenDB(userID)
}

func (m *Manager) createOrOpenDB(userID string) (*gorm.DB, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if db, exists := m.dbs[userID]; exists {
		return db, nil
	}

	if err := os.MkdirAll(m.dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(m.dataDir, fmt.Sprintf("%s.db", userID))

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database for user %s: %w", userID, err)
	}

	if err := m.runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations for user %s: %w", userID, err)
	}

	m.dbs[userID] = db
	return db, nil
}

func (m *Manager) runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&Post{},
		&Provider{},
		&ScheduledJob{},
		&APIToken{},
	)
}

func (m *Manager) CloseDB(userID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if db, exists := m.dbs[userID]; exists {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}

		if err := sqlDB.Close(); err != nil {
			return err
		}

		delete(m.dbs, userID)
	}

	return nil
}

func (m *Manager) CloseAll() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for userID, db := range m.dbs {
		sqlDB, err := db.DB()
		if err != nil {
			continue
		}

		if err := sqlDB.Close(); err != nil {
			continue
		}

		delete(m.dbs, userID)
	}

	return nil
}

func (m *Manager) GetDBPath(userID string) string {
	return filepath.Join(m.dataDir, fmt.Sprintf("%s.db", userID))
}

func (m *Manager) UserDBExists(userID string) bool {
	dbPath := m.GetDBPath(userID)
	_, err := os.Stat(dbPath)
	return err == nil
}

// GetAllUserDatabases returns all currently opened user databases
func (m *Manager) GetAllUserDatabases() map[string]*gorm.DB {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string]*gorm.DB)
	for userID, db := range m.dbs {
		result[userID] = db
	}

	return result
}

// NewTestManager creates a test database manager for testing
func NewTestManager(t *testing.T) *Manager {
	os.MkdirAll("./data", 0755) // Ensure ./data exists for test DBs
	tmpDir := "./data/test_socgo_" + time.Now().Format("20060102_150405")
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})
	return NewManager(tmpDir)
}

// Close closes all databases and cleans up resources
func (m *Manager) Close() error {
	return m.CloseAll()
}
