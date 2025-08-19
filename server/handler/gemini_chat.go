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
	completeHistory, err := curr_session.GetSessionHistory(ctx, sessionID, "complete")
	if err != nil {
		log.Printf("Error fetching session history: %v", err)
		return "Error fetching session history"
	}

	// Experimental Feature
	// A heuristic here to save on tokens. We will try to summarize the each 6 messages from the chat

	// We won't modify the complete history directly, instead we'll work with a copy of it.
	// With that copy, we will break apart the long context into smaller summaries that will save us tokens.
	summarizedHistory, err := curr_session.GetSessionHistory(ctx, sessionID, "summarized")
	if err != nil {
		log.Printf("Error fetching summarized session history: %v", err)
	}

	if len(completeHistory)%6 == 0 && len(completeHistory) > 0 {
		// If the complete history is a multiple of 6, we can summarize it
		for i := 0; i < len(completeHistory); i += 6 {
			end := i + 6
			if end > len(completeHistory) {
				// No need to summarize yet if the last chunk is not full
				break
			}

			// Summarize the chunk but only if they are not in the summarized history, we're doing so by a quick length comparison
			if len(summarizedHistory) > i/6 {
				continue
			}

			chunk := helper.CombineChatMessages(completeHistory[i:end])
			temp_history := []*genai.Content{genai.NewContentFromText(chunk, genai.RoleUser)}
			reply, _ := ai_models.GeminiChat(temp_history, "Please summarize the conversation without using pronouns and retain important details with at most 4 sentences. No need to double check with me.", model)
			summarizedHistory = append(summarizedHistory, template.Message{Content: reply, Role: genai.RoleModel})

			// Save the summarizedHistory to Redis under the "summarized" key
			if err := curr_session.SaveSessionHistory(ctx, sessionID, "summarized", summarizedHistory); err != nil {
				log.Printf("Error saving summarized session history: %v", err)
			}
		}
	}

	// Create a copy of the complete history to work with
	var geminiHistory []*genai.Content

	// Convert the existing messages stored in Redis to the format used by Gemini through genai.NewContentFromText
	// and add it to slice of genai.Content to be consume by the api call
	for _, msg := range completeHistory {
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
		lastChatIndexSummarized := len(summarizedHistory) * 6 // Very experimental. This gave us the lastest message index in the completeHistory that was summarized
		var combinedHistory []*genai.Content

		// This means chunk is not yet full.
		if a := len(completeHistory) - lastChatIndexSummarized; a < 6 && a > 0 {
			for _, msg := range summarizedHistory {
				combinedHistory = append(combinedHistory, genai.NewContentFromText(msg.Content, genai.Role(msg.Role)))
			}
			// Then we combine the summarized history with the complete history from the lastChatIndexSummarized to the end of geminiHistory
			combinedHistory = append(combinedHistory, geminiHistory[lastChatIndexSummarized+1:]...)
		} else {
			combinedHistory = geminiHistory
		}

		reply, updatedHistory := ai_models.GeminiChat(combinedHistory, input, model)

		// Convert the updated history back to the generic format
		for _, content := range updatedHistory {
			genericMsg := template.Message{Content: content.Parts[0].Text, Role: content.Role}
			completeHistory = append(completeHistory, genericMsg)
		}
		// Update the session history in Redis
		if err := curr_session.SaveSessionHistory(ctx, sessionID, "complete", completeHistory); err != nil {
			log.Printf("Error updating session history: %v", err)
		}

		return helper.HtmlOrCurlResponse(isHTML, reply)
	}
}
