package handler

import (
	"context"
	"fmt"
	"slices"

	"local_chatbot/internal/provider"
	"local_chatbot/server/ai_models"
	"local_chatbot/server/template"
	"local_chatbot/server/utility"

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
		utility.Logger.WithComponent("ollama_provider").Warn("Failed to connect to Ollama", "err", err)
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
	if slices.Contains(op.supportedModels, model) {
		op.selectedModel = model
		utility.Logger.WithComponent("ollama_provider").Debug("Set to", "model", op.selectedModel)
		return nil
	}
	return fmt.Errorf("model %s is not supported", model)
}

func (op *OllamaProvider) CompressHistory(history []template.Message) ([]template.Message, error) {
	// Because this is local, we can afford to keep more history context at the expense of responsiveness.
	if len(history) >= 20 {
		combinedContext := utility.CombineChatMessages(history[:10])

		ollamaHistory := []api.Message{
			{
				Role:    "user",
				Content: combinedContext,
			},
		}

		// Call Ollama to compress the history
		summarizedContext, err := ai_models.OllamaChat(
			ollamaHistory,
			op.Client,
			"Please summarize the conversation in a concise manner while preserving the key points.",
			"llama3.2:1b", // Use a smaller model for summarization to save resources
		)

		if err != nil {
			utility.Logger.WithComponent("ollama_provider").Error(err, "Error during Ollama summarization")
		}

		compressHistory := []template.Message{
			{
				Role:    "user",
				Content: summarizedContext.Content,
			},
		}

		compressHistory = append(compressHistory, history[10:]...)
		return compressHistory, nil
	}

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

	model := op.selectedModel

	resp, err := ai_models.OllamaChat(ollamaHistory, op.Client, userMessage, model)
	if err != nil {
		utility.Logger.WithComponent("ollama_provider").Error(err, "Error during Ollama chat")
		return "", err
	}

	// Extract the last response (which should be from the model)
	if resp.Content == "" {
		utility.Logger.WithComponent("ollama_provider").Warn("Received empty response from Ollama")
		return "", fmt.Errorf("Empty response from Ollama")
	}

	return utility.HtmlOrCurlResponse(isHTML, resp.Content), nil
}

// Close closes the Ollama provider
func (op *OllamaProvider) Close() error {
	// Ollama provider doesn't need explicit cleanup for now
	return nil
}
