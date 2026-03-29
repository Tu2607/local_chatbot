package app

import (
	"context"
	"fmt"
	"log"
	"sync"

	"local_chatbot/internal/config"
	"local_chatbot/internal/provider"
	"local_chatbot/server/handler"

	"github.com/redis/go-redis/v9"
)

// App holds all application dependencies
type App struct {
	Config           *config.Config
	ProviderRegistry *provider.Registry
	RedisClient      *redis.Client
	SessionManager   *handler.RedisSessionManager
	WgContextSync    sync.WaitGroup
}

// New initializes and returns a new App with all dependencies
func New(cfg *config.Config) (*App, error) {
	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: "",
		DB:       0,
	})

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	log.Println("Successfully connected to Redis")

	// Initialize session manager
	sessionManager := handler.NewRedisSessionManager(redisClient)

	// Initialize provider registry
	registry := provider.NewRegistry()

	// Initialize Gemini provider (only if API key is provided)
	if cfg.GeminiKey != "" {
		geminiProvider, err := initGeminiProvider(cfg.GeminiKey)
		if err != nil {
			log.Printf("Warning: Failed to initialize Gemini provider: %v", err)
		} else {
			registry.Register("gemini", geminiProvider)
			log.Println("Gemini provider initialized")
		}
	} else {
		log.Println("Skipping Gemini provider initialization (no API key configured)")
	}

	// Initialize Ollama provider (always attempt, it's local)
	ollamaProvider := initOllamaProvider()
	registry.Register("ollama", ollamaProvider)
	log.Println("Ollama provider initialized")

	return &App{
		Config:           cfg,
		ProviderRegistry: registry,
		RedisClient:      redisClient,
		SessionManager:   sessionManager,
		WgContextSync:    sync.WaitGroup{},
	}, nil
}

// Close gracefully shuts down all resources
func (a *App) Close() error {
	var lastErr error

	// Close provider registry (closes all providers)
	if err := a.ProviderRegistry.Close(); err != nil {
		log.Printf("Error closing providers: %v", err)
		lastErr = err
	}

	// Close Redis connection
	if err := a.RedisClient.Close(); err != nil {
		log.Printf("Error closing Redis: %v", err)
		lastErr = err
	}

	return lastErr
}

// Helper functions for provider initialization

func initGeminiProvider(apiKey string) (provider.Provider, error) {
	return handler.NewGeminiProvider(apiKey)
}

func initOllamaProvider() provider.Provider {
	return handler.NewOllamaProvider()
}

// WaitForContextSync waits for all background context operations to complete or for the context to be canceled
func (a *App) WaitForContextSync(ctx context.Context) error {
	finishCh := make(chan struct{})
	go func() {
		a.WgContextSync.Wait()
		close(finishCh)
	}()

	select {
	case <-finishCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
