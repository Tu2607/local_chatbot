package handler

import (
	"context"
	"log"

	"local_chatbot/server/ai_models"
	"local_chatbot/server/helper"
	"local_chatbot/server/template"

	"google.golang.org/genai"
)

func GeminiHandler(curr_session *RedisSessionManager, sessionID string, input string, model string, isHTML bool) string {
	ctx := context.Background()

	// Fetch the session history from redis, the returned value is a slice of generic
	// template struct of the message sent which contains the content and role.
	genericHistory, err := curr_session.GetSessionHistory(ctx, sessionID)
	if err != nil {
		log.Printf("Error fetching session history: %v", err)
		return "Error fetching session history"
	}

	// Convert the existing messages stored in Redis to the format used by Gemini through genai.NewContentFromText
	// and add it to slice of genai.Content to be consume by the api call
	var geminiHistory []*genai.Content
	for _, msg := range genericHistory {
		geminiHistory = append(geminiHistory, genai.NewContentFromText(msg.Content, genai.Role(msg.Role))) // Should have the correct role here
	}

	switch model {
	case "gemini-2.0-flash-preview-image-generation":
		// No need to update history for image generation, I'm assuming the context is not needed
		// if the prompt is sufficiently descriptive enough.
		// Well, maybe in the future when I have time to implement it.
		reply := ai_models.GeminiImageGeneration(input, model)
		return helper.EncodeByteSliceToBase64(reply)
	default:
		reply, updatedHistory := ai_models.GeminiChat(geminiHistory, input, model)
		// Convert the updated history back to the generic format
		for _, content := range updatedHistory {
			genericMsg := &template.Message{Content: content.Parts[0].Text, Role: content.Role}
			genericHistory = append(genericHistory, genericMsg)
		}
		// Update the session history in Redis
		if err := curr_session.SaveSessionHistory(ctx, sessionID, genericHistory); err != nil {
			log.Printf("Error updating session history: %v", err)
		}

		return helper.HtmlOrCurlResponse(isHTML, reply)
	}
}
