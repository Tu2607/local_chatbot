package ai_models

import (
	"context"
	"log"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func OpenAIChat(prompt string, model string) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is not set")
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))
	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Say this is a test"),
		},
		Model: openai.ChatModelGPT4o,
	})

	if err != nil {
		log.Fatalf("Failed to create chat completion: %v", err)
	}

	println("Chat completion response:", chatCompletion.Choices[0].Message.Content)
}
