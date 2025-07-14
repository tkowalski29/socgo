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