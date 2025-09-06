package ai_models

import (
	"context"
	"log"

	"github.com/ollama/ollama/api"
)

func OllamaGenerateText(prompt string, model string) (string, error) {
	// Primarily will be use to summarize text
	// Placeholder for Ollama text generation logic
	// This function should implement the logic to interact with the Ollama model
	// and return the generated text based on the provided prompt and model.
	return "", nil
}

func OllamaChat(history []api.Message, prompt string, model string) []api.Message {
	// Placeholder for Ollama chat logic
	// This function should implement the logic to interact with the Ollama model
	// and return the response based on the provided prompt and model.
	ctx := context.Background()

	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	history = append(history, api.Message{
		Role:    "user",
		Content: prompt,
	})

	new_req := &api.ChatRequest{
		Model:    model,
		Messages: history,
		Stream:   func() *bool { b := false; return &b }(), // Enable/Disable streaming. Set to false for now.
	}

	err = client.Chat(ctx, new_req, func(cr api.ChatResponse) error {
		chat_resp := api.Message{
			Role:    cr.Message.Role,
			Content: cr.Message.Content,
		}
		history = append(history, chat_resp)
		return nil
	})

	if err != nil {
		log.Fatalf("Failed to get response: %v", err)
	}

	return history
}
