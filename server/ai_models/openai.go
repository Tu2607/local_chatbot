package ai_models

import (
	"context"
	"log"

	"local_chatbot/server/template"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/responses"
)

func OpenAIText(client *openai.Client, history []template.Message, prompt string, model string) string {
	// There is a method that return a conversation object that return an object with the conversation context and its ID.
	//
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
