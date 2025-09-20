package handler

import (
	"context"
	"local_chatbot/server/ai_models"
	"local_chatbot/server/helper"
	"local_chatbot/server/template"
	"log"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

type OpenAIClient struct {
	Client          *openai.Client
	SupportedModels []string
}

func NewOpenAIClient(api_key string) OpenAIClient {
	client := openai.NewClient(option.WithAPIKey(api_key))
	return OpenAIClient{
		Client:          &client,
		SupportedModels: []string{"gpt-4o", "gpt-4.1-mini"},
	}
}

func (client *OpenAIClient) Chat(history []template.Message, prompt string, model string) string {
	return ai_models.OpenAIText(client.Client, history, prompt, model)
}

// OpenAIHandler handles the OpenAI chat requests, will not have context saving for now
// since implementing session management with Redis is planned for the future.
func (client *OpenAIClient) OpenAIHandler(curr_session *RedisSessionManager, sessionID string, input string, model string, isHTML bool) string {
	ctx := context.Background()
	// Fetch the session history from redis
	history, err := curr_session.GetSessionHistory(ctx, sessionID, "complete")
	if err != nil {
		log.Printf("Error fetching session history: %v", err)
	}

	switch model {
	case "gpt-4.1-mini":
		reply := client.Chat(history, input, model)
		reply = helper.HtmlOrCurlResponse(isHTML, reply)
		return reply
	default:
		return "Unsupported OpenAI model: " + model
	}
}
