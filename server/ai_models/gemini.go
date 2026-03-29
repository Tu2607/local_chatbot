package ai_models

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

func GeminiChat(chat_ctx []*genai.Content, client *genai.Client, prompt string, selected_model string) (string, error) {
	ctx := context.Background()

	latest_user_msg := []*genai.Content{
		genai.NewContentFromText(prompt, genai.RoleUser),
	}

	// Append the history to the contents
	latest_chat_ctx := append(chat_ctx, latest_user_msg...)

	result, err := client.Models.GenerateContent(
		ctx,
		selected_model,
		latest_chat_ctx,
		nil, // Could add additional parameters here if needed for more thinking tokens
	)

	if err != nil {
		return "", err
	}

	var resultText string

	if len(result.Candidates) > 0 {
		resultText = result.Candidates[0].Content.Parts[0].Text
	} else {
		return "", fmt.Errorf("no candidates returned from Gemini API")
	}

	return resultText, nil
}

// The method definition is a bit future proofing as gemini-2.0-flash-preview-image-generation
// is the only model that supports image generation at the moment. More models may support this in the future.
// Hence we keep the string parameter for the model name.
func GeminiImageGeneration(prompt string, client *genai.Client, selected_model string) ([]byte, error) {
	ctx := context.Background()

	config := &genai.GenerateContentConfig{
		ResponseModalities: []string{"TEXT", "IMAGE"},
	}

	result, err := client.Models.GenerateContent(
		ctx,
		selected_model,
		genai.Text(prompt),
		config,
	)

	if err != nil {
		return nil, err
	}

	var imageData []byte
	for _, part := range result.Candidates[0].Content.Parts {
		if part.InlineData != nil {
			imageData = part.InlineData.Data
		}
	}

	return imageData, nil
}
