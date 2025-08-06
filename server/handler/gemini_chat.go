package handler

import (
	"local_chatbot/server/ai_models"
	"local_chatbot/server/helper"
)

func GeminiHandler(sessionID string, input string, model string, isHTML bool) string {
	genai_sessions.RLock()
	history := genai_sessions.histories[sessionID]
	genai_sessions.RUnlock()

	switch model {
	case "gemini-2.0-flash-preview-image-generation":
		// No need to update history for image generation, I'm assuming the context is not needed
		// if the prompt is sufficiently descriptive enough.
		// Well, maybe in the future when I have time to implement it.
		reply := ai_models.GeminiImageGeneration(input, model)
		encodedReply := helper.EncodeByteSliceToBase64(reply)
		return encodedReply
	default:
		reply, updatedHistory := ai_models.GeminiChat(history, input, model)
		genai_sessions.Lock()
		genai_sessions.histories[sessionID] = updatedHistory // Update the session history atomically
		genai_sessions.Unlock()
		return helper.HtmlOrCurlResponse(isHTML, reply)
	}
}
