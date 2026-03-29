package provider

import (
	"context"

	"local_chatbot/server/template"
)

// Provider is the unified interface for all LLM providers
type Provider interface {
	// SendMessage sends a message to the LLM and returns the response
	SendMessage(ctx context.Context, sessionID string, userMessage string, history []template.Message, isHTML bool) (response string, err error)

	// GetSupportedModels returns the list of models supported by this provider
	GetSupportedModels() []string

	// Set selected model for the provider, if applicable (e.g., for providers that support multiple models)
	SetModel(model string) error

	// CompressHistory compresses the session history to preserve context for long conversations while staying within token limits
	CompressHistory(history []template.Message) ([]template.Message, error)

	// Name returns the provider name (e.g., "gemini", "ollama")
	Name() string

	// Close closes any resources used by the provider
	Close() error
}
