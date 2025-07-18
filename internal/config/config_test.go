package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	config, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if config.Server.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", config.Server.Port)
	}

	if config.Server.Host != "localhost" {
		t.Errorf("Expected default host localhost, got %s", config.Server.Host)
	}
}

func TestGetServerAddr(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: "8080",
		},
	}

	expected := "localhost:8080"
	if addr := config.GetServerAddr(); addr != expected {
		t.Errorf("Expected %s, got %s", expected, addr)
	}
}

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	if value := getEnv("TEST_VAR", "default"); value != "test_value" {
		t.Errorf("Expected test_value, got %s", value)
	}

	if value := getEnv("NON_EXISTENT", "default"); value != "default" {
		t.Errorf("Expected default, got %s", value)
	}
}

func TestDetectNgrokURL(t *testing.T) {
	// Test when ngrok is not running (should return empty string)
	// Note: This test may pass or fail depending on whether ngrok is running
	url := detectNgrokURL()
	// We can't reliably test this without mocking, so we'll just check it doesn't panic
	if url != "" {
		t.Logf("Ngrok URL detected: %s (this is expected if ngrok is running)", url)
	}
}

func TestLoadFromEnvWithNgrokURL(t *testing.T) {
	// Set NGROK_URL environment variable
	os.Setenv("NGROK_URL", "https://test.ngrok.io")
	defer os.Unsetenv("NGROK_URL")

	config := loadFromEnv()
	expected := "https://test.ngrok.io"
	if config.Server.BaseURL != expected {
		t.Errorf("Expected BaseURL %s, got %s", expected, config.Server.BaseURL)
	}
}

func TestLoadFromEnvWithServerBaseURL(t *testing.T) {
	// Set SERVER_BASE_URL environment variable (should take precedence)
	os.Setenv("SERVER_BASE_URL", "https://custom.example.com")
	defer os.Unsetenv("SERVER_BASE_URL")

	config := loadFromEnv()
	expected := "https://custom.example.com"
	if config.Server.BaseURL != expected {
		t.Errorf("Expected BaseURL %s, got %s", expected, config.Server.BaseURL)
	}
}

func TestLoadFromEnvWithServerBaseURLPrecedence(t *testing.T) {
	// Test that SERVER_BASE_URL takes precedence over NGROK_URL
	os.Setenv("SERVER_BASE_URL", "https://custom.example.com")
	os.Setenv("NGROK_URL", "https://test.ngrok.io")
	defer func() {
		os.Unsetenv("SERVER_BASE_URL")
		os.Unsetenv("NGROK_URL")
	}()

	config := loadFromEnv()
	expected := "https://custom.example.com"
	if config.Server.BaseURL != expected {
		t.Errorf("Expected BaseURL %s, got %s", expected, config.Server.BaseURL)
	}
}
