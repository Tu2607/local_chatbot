package ai_models

import (
	"context"
	"fmt"

	"local_chatbot/server/template"
	"local_chatbot/server/utility"

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

func GeminiEmbedText(client *genai.Client, text string) ([]float32, error) {
	ctx := context.Background()

	// A sanitize step to check if the input text is base64 encoded, if so we decode it before sending to Gemini API for embedding generation.
	// This is to handle the case where the input text is actually an image in base64 format,
	// which can happen when the frontend sends an image as a base64 string for embedding generation.
	if utility.IsBase64Encoded(text) {
		decodedBytes, err := utility.DecodeBase64ToByteSlice(text)
		if err != nil {
			utility.Logger.WithComponent("gemini_embedding").Error(err, "Error decoding base64 text:", "text", text)
			return nil, err
		}
		text = string(decodedBytes)
	}

	// Convert the text to Gemini format
	var content []*genai.Content
	content = append(content, genai.NewContentFromText(text, genai.RoleUser))

	result, err := client.Models.EmbedContent(
		ctx,
		"gemini-embedding-2-preview",
		content,
		nil,
	)
	if err != nil {
		utility.Logger.WithComponent("gemini_embedding").Error(err, "Error generating embedding for text:", "text", text)
		return nil, err
	}

	if len(result.Embeddings) > 0 {
		return result.Embeddings[0].Values, nil
	}

	return nil, fmt.Errorf("No embeddings returned from Gemini API")
}

func GeminiEmbedTextsFromHistory(client *genai.Client, history []template.Message) ([][]float32, error) {
	var embeddedVectors [][]float32
	// Go through each message in the history
	for _, msg := range history {
		embedding, err := GeminiEmbedText(client, msg.Content)
		if err != nil {
			utility.Logger.WithComponent("gemini_embedding_wrapper").Error(err, "Error generating embedding for message:", "message", msg.Content)
			return nil, err
		}
		embeddedVectors = append(embeddedVectors, embedding)
	}
	return embeddedVectors, nil
}
