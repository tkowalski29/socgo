package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	DB        DBConfig        `yaml:"db"`
	Database  DatabaseConfig  `yaml:"database"`
	Providers ProvidersConfig `yaml:"providers"`
}

type ServerConfig struct {
	Port    string `yaml:"port"`
	Host    string `yaml:"host"`
	BaseURL string `yaml:"base_url"`
}

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

type DatabaseConfig struct {
	DataDir string `yaml:"data_dir"`
}

type ProvidersConfig struct {
	TikTok    []ProviderInstance `yaml:"tiktok"`
	Instagram []ProviderInstance `yaml:"instagram"`
	Facebook  []ProviderInstance `yaml:"facebook"`
}

type ProviderInstance struct {
	Name         string `yaml:"name"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	Description  string `yaml:"description,omitempty"`
}

func Load() (*Config, error) {
	loadEnvFile()

	// Try to load from config.yml first
	if config, err := loadFromYAML(); err == nil {
		return config, nil
	}

	// Fallback to environment variables
	log.Println("No config.yml found, using environment variables")
	return loadFromEnv(), nil
}

func loadFromYAML() (*Config, error) {
	data, err := os.ReadFile("config.yml")
	if err != nil {
		return nil, fmt.Errorf("failed to read config.yml: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config.yml: %w", err)
	}

	// Set defaults for missing values
	setDefaults(&config)

	return &config, nil
}

func loadFromEnv() *Config {
	baseURL := getEnv("SERVER_BASE_URL", "")
	if baseURL == "" {
		// Check for NGROK_URL environment variable first
		if ngrokURL := getEnv("NGROK_URL", ""); ngrokURL != "" {
			baseURL = ngrokURL
			log.Printf("Using NGROK_URL from environment: %s", ngrokURL)
		} else {
			// Try to detect ngrok URL automatically
			if ngrokURL := detectNgrokURL(); ngrokURL != "" {
				baseURL = ngrokURL
				log.Printf("Auto-detected ngrok URL: %s", ngrokURL)
			} else {
				baseURL = "http://localhost:8080"
			}
		}
	}

	return &Config{
		Server: ServerConfig{
			Port:    getEnv("SERVER_PORT", "8080"),
			Host:    getEnv("SERVER_HOST", "localhost"),
			BaseURL: baseURL,
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "socgo"),
		},
		Database: DatabaseConfig{
			DataDir: getEnv("DATABASE_DATA_DIR", "./data"),
		},
		Providers: ProvidersConfig{
			TikTok:    []ProviderInstance{},
			Instagram: []ProviderInstance{},
			Facebook:  []ProviderInstance{},
		},
	}
}

func setDefaults(config *Config) {
	if config.Server.Port == "" {
		config.Server.Port = "8080"
	}
	if config.Server.Host == "" {
		config.Server.Host = "localhost"
	}
	if config.Server.BaseURL == "" {
		// Try to detect ngrok URL automatically
		if ngrokURL := detectNgrokURL(); ngrokURL != "" {
			config.Server.BaseURL = ngrokURL
			log.Printf("Auto-detected ngrok URL: %s", ngrokURL)
		} else {
			config.Server.BaseURL = "http://localhost:8080"
		}
	}
	if config.Database.DataDir == "" {
		config.Database.DataDir = "./data"
	}
}

func loadEnvFile() {
	file, err := os.Open(".env")
	if err != nil {
		log.Println("No .env file found, using defaults and environment variables")
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Failed to close .env file: %v", err)
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}

// GetProviderConfig returns the provider instance by name and type
func (c *Config) GetProviderConfig(providerType, name string) (*ProviderInstance, error) {
	var instances []ProviderInstance

	switch providerType {
	case "tiktok":
		instances = c.Providers.TikTok
	case "instagram":
		instances = c.Providers.Instagram
	case "facebook":
		instances = c.Providers.Facebook
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}

	for _, instance := range instances {
		if instance.Name == name {
			return &instance, nil
		}
	}

	return nil, fmt.Errorf("provider instance not found: %s/%s", providerType, name)
}

// GetAllProviderInstances returns all provider instances for a given type
func (c *Config) GetAllProviderInstances(providerType string) []ProviderInstance {
	switch providerType {
	case "tiktok":
		return c.Providers.TikTok
	case "instagram":
		return c.Providers.Instagram
	case "facebook":
		return c.Providers.Facebook
	default:
		return []ProviderInstance{}
	}
}

// detectNgrokURL tries to detect ngrok URL from ngrok API
func detectNgrokURL() string {
	// Try ngrok API endpoints
	apiEndpoints := []string{
		"http://localhost:4040/api/tunnels",
		"http://127.0.0.1:4040/api/tunnels",
	}

	for _, endpoint := range apiEndpoints {
		if url := tryNgrokAPI(endpoint); url != "" {
			return url
		}
	}

	return ""
}

// tryNgrokAPI attempts to get ngrok URL from the API
func tryNgrokAPI(endpoint string) string {
	client := &http.Client{Timeout: 2 * time.Second}

	resp, err := client.Get(endpoint)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var result struct {
		Tunnels []struct {
			PublicURL string `json:"public_url"`
			Proto     string `json:"proto"`
		} `json:"tunnels"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}

	// Find the first HTTPS tunnel
	for _, tunnel := range result.Tunnels {
		if tunnel.Proto == "https" && tunnel.PublicURL != "" {
			return tunnel.PublicURL
		}
	}

	// If no HTTPS, return the first available tunnel
	if len(result.Tunnels) > 0 && result.Tunnels[0].PublicURL != "" {
		return result.Tunnels[0].PublicURL
	}

	return ""
}
