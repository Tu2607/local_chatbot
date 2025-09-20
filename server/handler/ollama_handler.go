package handler

import (
	"context"
	"log"

	"local_chatbot/server/ai_models"
	"local_chatbot/server/helper"
	"local_chatbot/server/template"

	"github.com/ollama/ollama/api"
)

type OllamaClient struct {
	Client          *api.Client
	SupportedModels []string
}

// Constructor for OllamaClient
func NewOllamaClient() OllamaClient {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatalf("Failed to create Ollama client: %v", err)
	} else {
		log.Println("Successfully created Ollama client")
	}
	return OllamaClient{
		Client: client,
		SupportedModels: []string{
			"llama3.2:1b",
			"qwen3-coder:30b",
			"gpt-oss:20b",
			"gemma3:27b",
		},
	}
}

func (client *OllamaClient) Chat(history []api.Message, prompt string, selected_model string) []api.Message {
	return ai_models.OllamaChat(history, client.Client, prompt, selected_model)
}

func (client *OllamaClient) OllamaHandler(curr_session *RedisSessionManager, sessionID string, input string, model string, isHTML bool) string {
	// Place holder for Ollama support
	ctx := context.Background()

	// Fetch the session history from redis, the returned value is a slice of generic
	// template struct of the message sent which contains the content and role.
	completeHistory, err := curr_session.GetSessionHistory(ctx, sessionID, "complete")
	if err != nil {
		log.Printf("Error fetching session history: %v", err)
		return "Error fetching session history"
	}

	// Probably will implement some summarization logic here later to save on context length for Ollama models

	// Because Ollama API expects a slice of api.Message, we need to convert our template.Message slice to api.Message slice
	var ollamaHistory []api.Message
	for _, msg := range completeHistory {
		ollamaHistory = append(ollamaHistory, api.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Make a call to Ollama chat function
	ollamaHistory = client.Chat(ollamaHistory, input, model)

	// Append the latest user input and the response to the complete history
	completeHistory = append(completeHistory, template.Message{Content: input, Role: "user"})
	// The last message in ollamaHistory is the model's response
	latestResponse := ollamaHistory[len(ollamaHistory)-1]
	completeHistory = append(completeHistory, template.Message{Content: latestResponse.Content, Role: latestResponse.Role})

	// Save the updated complete history back to Redis
	err = curr_session.SaveSessionHistory(ctx, sessionID, "complete", completeHistory)
	if err != nil {
		log.Printf("Error saving session history: %v", err)
		return "Error saving session history"
	}

	// Lock in the model used for this session if not already set
	if err := curr_session.SaveSessionModel(ctx, sessionID, model); err != nil {
		log.Printf("Error setting model for session: %v", err)
	}

	return helper.HtmlOrCurlResponse(isHTML, latestResponse.Content)
}
