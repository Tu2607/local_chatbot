package ai_models

import (
	"context"
	"log"
	"os"

	"google.golang.org/genai"
)

func GeminiChat(history []*genai.Content, prompt string, selected_model string) (string, []*genai.Content) {
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

	// Append the history to the contents
	updatedHistory := append(history, contents...)

	result, err := client.Models.GenerateContent(
		ctx,
		selected_model,
		updatedHistory,
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
		log.Fatalf("No candidates returned from the model")
	}

	// Append the response to the history
	updatedHistory = append(updatedHistory, genai.NewContentFromText(resultText, genai.RoleModel))

	return resultText, updatedHistory
}

// The method definition is a bit future proofing as gemini-2.0-flash-preview-image-generation
// is the only model that supports image generation at the moment. More models may support this in the future.
// Hence we keep the string parameter for the model name.
func GeminiImageGeneration(prompt string, selected_model string) []byte {
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

	config := &genai.GenerateContentConfig{
		ResponseModalities: []string{"TEXT", "IMAGE"},
	}

	result, _ := client.Models.GenerateContent(
		ctx,
		selected_model,
		genai.Text(prompt),
		config,
	)

	var imageURL []byte
	for _, part := range result.Candidates[0].Content.Parts {
		if part.InlineData != nil {
			imageURL = part.InlineData.Data
		}
	}

	return imageURL
}
