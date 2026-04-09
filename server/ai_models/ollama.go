package ai_models

import (
	"context"

	"local_chatbot/server/utility"

	"github.com/ollama/ollama/api"
)

func OllamaChat(history []api.Message, client *api.Client, prompt string, model string) (api.Message, error) {
	ctx := context.Background()

	history = append(history, api.Message{
		Role:    "user",
		Content: prompt,
	})

	new_req := &api.ChatRequest{
		Model:    model,
		Messages: history,
		Stream:   func() *bool { b := false; return &b }(), // Enable/Disable streaming. Set to false for now.
	}
	var chat_resp api.Message
	err := client.Chat(ctx, new_req, func(cr api.ChatResponse) error {
		chat_resp = api.Message{
			Role:    cr.Message.Role,
			Content: cr.Message.Content,
		}
		return nil
	})

	if err != nil {
		utility.Logger.WithComponent("ollama_chat").Error(err, "Error during Ollama chat")
		return api.Message{}, err
	}

	return chat_resp, nil
}
