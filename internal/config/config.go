package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"local_chatbot/server/utility"
)

// Config holds all application configuration
type Config struct {
	Port      string
	RedisAddr string
	RedisPort string
	GeminiKey string
	DataDir   string // Directory for storing documents and RAG data

	// Below are some Ollama specific configurations
	OllamaHost string
}

// Load loads configuration from environment variables with defaults
func Load() (*Config, error) {
	cfg := &Config{
		Port:       getEnv("PORT", "55572"),
		RedisAddr:  getEnv("REDIS_ADDR", "localhost"),
		RedisPort:  getEnv("REDIS_PORT", "6379"),
		GeminiKey:  getEnv("GEMINI_API_KEY", ""),
		DataDir:    getEnv("DATA_DIR", "./data"),
		OllamaHost: getEnv("OLLAMA_HOST", "ollama:11434"),
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
		utility.Logger.WithComponent("config").Warn("GEMINI_API_KEY is not set. Gemini provider will not be available.")
	}

	// Validate Redis port is a number
	if _, err := strconv.Atoi(c.RedisPort); err != nil {
		return fmt.Errorf("REDIS_PORT must be a valid number, got: %s", c.RedisPort)
	}

	// Validate application port is a number
	if _, err := strconv.Atoi(c.Port); err != nil {
		return fmt.Errorf("PORT must be a valid number, got: %s", c.Port)
	}

	// Validate Ollama's port is a number
	OllamaPort := strings.Split(c.OllamaHost, ":")[1]
	if _, err := strconv.Atoi(OllamaPort); err != nil {
		return fmt.Errorf("Ollama's port must be a valid number, got: %s", OllamaPort)
	}

	return nil
}

// GetRedisAddr returns the full Redis address (host:port)
func (c *Config) GetRedisAddr() string {
	return c.RedisAddr + ":" + c.RedisPort
}

// Get jus the host part of the Ollama address (without port)
func (c *Config) GetOllamaHost() string {
	return strings.Split(c.OllamaHost, ":")[0]
}

// Helper function to get environment variable with default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
