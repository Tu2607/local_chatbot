package ai_models

import (
	"context"
	"log"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/responses"
)

func OpenAIText(prompt string, model string) string {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is not set")
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))
	// Instead of using chat completions, we use the response API directly
	resp_params := responses.ResponseNewParams{
		Model: model,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(prompt),
		},
		Temperature: openai.Float(0.7), // Messing around with temperature for creative responses, may cause hallucinations.
		// PreviousResponseID: nil, // This will be used for context in future implementations with Redis
	}

	resp, err := client.Responses.New(context.Background(), resp_params)
	if err != nil {
		log.Fatalf("Failed to create chat completion: %v", err)
	}

	reply := resp.OutputText()
	return reply
}
