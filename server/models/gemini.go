package models

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/genai"
)

func geminiChat(prompt string, selected_model string) *genai.GenerateContentResponse {
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

	fmt.Println("Generated content:", result)
	return result
}

func main() {
	test := geminiChat("Hello, how are you?", "gemini-2.5-flash")
	fmt.Println("Test result:", test)
}
