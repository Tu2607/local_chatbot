package ai_models

import (
	"context"
	"log"
	"os"

	"google.golang.org/genai"
)

func GeminiChat(prompt string, selected_model string) string {
	ctx := context.Background()

	// Initialize the GenAI client with the API key from environment variable
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable is not set")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})

	if err != nil {
		log.Fatalf("Failed to create GenAI client: %v", err)
	}

	contents := []*genai.Content{
		genai.NewContentFromText(prompt, genai.RoleUser),
	}

	result, err := client.Models.GenerateContent(
		ctx,
		selected_model,
		contents,
		nil, // Could add additional parameters here if needed for more thinking tokens
	)

	if err != nil {
		log.Fatalf("Failed to generate content: %v", err)
	}

	// For now, we only support text responses so we return the first candidate
	var resultText string

	if len(result.Candidates) > 0 {
		resultText = result.Candidates[0].Content.Parts[0].Text
	} else {
		log.Println("No candidates returned from the model")
	}

	return resultText
}
