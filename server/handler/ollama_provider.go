package handler

import (
	"context"
	"fmt"
	"log"

	"local_chatbot/internal/provider"
	"local_chatbot/server/ai_models"
	"local_chatbot/server/helper"
	"local_chatbot/server/template"

	"github.com/ollama/ollama/api"
)

// OllamaProvider implements the provider.Provider interface for Ollama
type OllamaProvider struct {
	Client          *api.Client
	supportedModels []string
	selectedModel   string
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider() provider.Provider {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Printf("Warning: Failed to connect to Ollama: %v", err)
	}

	return &OllamaProvider{
		Client: client,
		supportedModels: []string{
			"llama3.2:1b",
			"qwen3-coder:30b",
			"gpt-oss:20b",
			"gemma3:27b",
		},
	}
}

// Name returns the provider name
func (op *OllamaProvider) Name() string {
	return "ollama"
}

// GetSupportedModels returns the list of supported models
func (op *OllamaProvider) GetSupportedModels() []string {
	return op.supportedModels
}

func (op *OllamaProvider) SetModel(model string) error {
	for _, m := range op.supportedModels {
		if m == model {
			op.selectedModel = model
			return nil
		}
	}
	return fmt.Errorf("model %s is not supported", model)
}

func (op *OllamaProvider) CompressHistory(history []template.Message) ([]template.Message, error) {
	// For simplicity, we will just return the last 10 messages in the history.
	// In a real implementation, you would want to use a more sophisticated approach to compress the history while preserving important context.
	return history, nil
}

// SendMessage sends a message to Ollama and returns the response
func (op *OllamaProvider) SendMessage(ctx context.Context, sessionID string, userMessage string, history []template.Message, isHTML bool) (string, error) {
	// Convert history to Ollama format
	var ollamaHistory []api.Message
	for _, msg := range history {
		ollamaHistory = append(ollamaHistory, api.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Use the first supported model as default
	model := op.supportedModels[0] // llama3.2:1b as default

	// Call Ollama chat
	updatedHistory := ai_models.OllamaChat(ollamaHistory, op.Client, userMessage, model)

	// Extract the last response (which should be from the model)
	if len(updatedHistory) == 0 {
		return "", fmt.Errorf("no response from Ollama")
	}

	lastMessage := updatedHistory[len(updatedHistory)-1]
	return helper.HtmlOrCurlResponse(isHTML, lastMessage.Content), nil
}

// Close closes the Ollama provider
func (op *OllamaProvider) Close() error {
	// Ollama provider doesn't need explicit cleanup for now
	return nil
}
