package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	Port      string
	RedisAddr string
	RedisPort string
	GeminiKey string
	DataDir   string // Directory for storing documents and RAG data
}

// Load loads configuration from environment variables with defaults
func Load() (*Config, error) {
	cfg := &Config{
		Port:      getEnv("PORT", "55572"),
		RedisAddr: getEnv("REDIS_ADDR", "localhost"),
		RedisPort: getEnv("REDIS_PORT", "6379"),
		GeminiKey: getEnv("GEMINI_API_KEY", ""),
		DataDir:   getEnv("DATA_DIR", "./data"),
	}

	// Validate required configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks that required configuration is present
func (c *Config) Validate() error {
	// Warn if Gemini API key is not set (optional for now - allows development without keys)
	if c.GeminiKey == "" {
		log.Println("Warning: GEMINI_API_KEY is not set. Gemini provider will not be available.")
	}

	// Validate Redis port is a number
	if _, err := strconv.Atoi(c.RedisPort); err != nil {
		return fmt.Errorf("REDIS_PORT must be a valid number, got: %s", c.RedisPort)
	}

	// Validate application port is a number
	if _, err := strconv.Atoi(c.Port); err != nil {
		return fmt.Errorf("PORT must be a valid number, got: %s", c.Port)
	}

	return nil
}

// GetRedisAddr returns the full Redis address (host:port)
func (c *Config) GetRedisAddr() string {
	return c.RedisAddr + ":" + c.RedisPort
}

// Helper function to get environment variable with default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
