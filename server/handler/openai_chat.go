package handler

import (
	"local_chatbot/server/ai_models"
	"local_chatbot/server/helper"
)

// OpenAIHandler handles the OpenAI chat requests, will not have context saving for now
// since implementing session management with Redis is planned for the future.
func OpenAIHandler(input string, model string, isHTML bool) string {
	// Call the OpenAI chat function and get the response
	switch model {
	case "gpt-4.1-mini":
		reply := ai_models.OpenAIText(input, model)
		reply = helper.HtmlOrCurlResponse(isHTML, reply)
		return reply
	default:
		return "Unsupported OpenAI model: " + model
	}
}
