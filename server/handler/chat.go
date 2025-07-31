package handler

import (
	"encoding/json"
	"local_chatbot/server/ai_models"
	"local_chatbot/server/helper"
	"net/http"
)

type ChatRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

type ChatResponse struct {
	Response string `json:"response"`
}

func ChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method. Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	// Decode the JSON request body, filling out the input and model fields
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	// Check if the response should be in HTML format
	isHTML := r.URL.Query().Get("format") == "html"

	// TODO: Implement chat logic here
	var reply string

	switch req.Model {
	case "gemini-2.5-flash":
		reply = ai_models.GeminiChat(req.Input, "gemini-2.5-flash")
	case "gemini-2.5-pro":
		reply = ai_models.GeminiChat(req.Input, "gemini-2.5-pro")
	case "gemini-2.5-flash-lite": // The cheapest model
		reply = ai_models.GeminiChat(req.Input, "gemini-2.5-flash-lite")
	case "gemini-2.0-flash-preview-image-generation":
		reply = ai_models.GeminiImageGeneration(req.Input, "gemini-2.0-flash-preview-image-generation")
	case "local":
		reply = ai_models.OllamaChat(req.Input, "")
	}

	// Parse the response in a way that is readable on the Web UI and at the same time
	// acceptable for curl
	if isHTML {
		htmlOutput, err := helper.MarkdownToHTML(reply)
		if err != nil {
			http.Error(w, "Failed to convert Markdown to HTML", http.StatusInternalServerError)
			return
		}
		reply = htmlOutput
	}

	resp := ChatResponse{Response: reply}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
