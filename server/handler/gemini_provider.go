package handler

import (
	"context"
	"fmt"
	"log"
	"slices"

	"local_chatbot/internal/provider"
	"local_chatbot/server/ai_models"
	"local_chatbot/server/helper"
	"local_chatbot/server/template"

	"google.golang.org/genai"
)

// GeminiProvider implements the provider.Provider interface for Gemini
type GeminiProvider struct {
	Client          *genai.Client
	supportedModels []string
	selectedModel   string
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(apiKey string) (provider.Provider, error) {
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}

	log.Println("Successfully created Gemini client")
	return &GeminiProvider{
		Client: client,
		supportedModels: []string{
			"gemma-3-27b-it",
			"gemini-2.5-flash",
			"gemini-2.5-pro",
			"gemini-2.5-flash-lite",
			"gemini-2.0-flash-preview-image-generation",
		},
	}, nil
}

// Name returns the provider name
func (gp *GeminiProvider) Name() string {
	return "gemini"
}

func (gp *GeminiProvider) SetModel(model string) error {
	// For simplicity, we will just check if the model is in the supported models list and return an error if it's not.
	// In a real implementation, you would want to have a more robust way of handling model selection and validation.
	if slices.Contains(gp.supportedModels, model) {
		gp.selectedModel = model
		return nil
	}
	gp.selectedModel = "gemini-2.5-flash-lite" // Default to a known model if the provided one is not supported
	return fmt.Errorf("model %s is not supported by Gemini provider. Defaulting to %s", model, gp.selectedModel)
}

func (gp *GeminiProvider) CompressHistory(history []template.Message) ([]template.Message, error) {
	// Select the oldest 10 messages to compress the history, this is a simple heuristic and can be improved with more sophisticated approaches
	if len(history) > 10 {
		combinedContent := helper.CombineChatMessages(history[:10])
		geminiHistory := []*genai.Content{
			genai.NewContentFromText(combinedContent, genai.RoleUser),
		}

		summarizedContent, err := ai_models.GeminiChat(
			geminiHistory,
			gp.Client,
			"Please summarize the conversation in a concise manner while preserving the key points.",
			"gemini-2.5-flash-lite", // Use a smaller model for summarization to save tokens
		)

		if err != nil {
			log.Printf("Warning: Could not compress history: %v. Continuing with uncompressed.", err)
			return history, nil
		}

		compressedHistory := []template.Message{
			{
				Role:    "user",
				Content: summarizedContent,
			},
		}
		compressedHistory = append(compressedHistory, history[10:]...)
		return compressedHistory, nil
	}

	// If the history is 10 messages or less, we can just return it as is
	return history, nil
}

// GetSupportedModels returns the list of supported models
func (gp *GeminiProvider) GetSupportedModels() []string {
	return gp.supportedModels
}

// SendMessage sends a message to Gemini and returns the response
func (gp *GeminiProvider) SendMessage(ctx context.Context, sessionID string, userMessage string, history []template.Message, isHTML bool) (string, error) {
	// Handle image generation separately
	if gp.selectedModel == "gemini-2.0-flash-preview-image-generation" {
		reply, err := ai_models.GeminiImageGeneration(userMessage, gp.Client, gp.selectedModel)
		if err != nil {
			log.Printf("Error generating image with Gemini: %v", err)
			return "Failed to generate image", nil // Return a user-friendly message instead of an error
		}
		return helper.EncodeByteSliceToBase64(reply), nil
	}

	// Convert history to Gemini format
	var geminiHistory []*genai.Content
	for _, msg := range history {
		geminiHistory = append(geminiHistory, genai.NewContentFromText(msg.Content, genai.Role(msg.Role)))
	}

	contents := []*genai.Content{
		genai.NewContentFromText(userMessage, genai.RoleUser),
	}

	updatedHistory := append(geminiHistory, contents...)

	result, err := gp.Client.Models.GenerateContent(
		ctx,
		gp.selectedModel,
		updatedHistory,
		nil,
	)
	if err != nil {
		return "", err
	}

	var resultText string
	if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
		resultText = result.Candidates[0].Content.Parts[0].Text
	}

	return helper.HtmlOrCurlResponse(isHTML, resultText), nil
}

// Close closes the Gemini provider
func (gp *GeminiProvider) Close() error {
	// Gemini client doesn't require explicit cleanup
	return nil
}
