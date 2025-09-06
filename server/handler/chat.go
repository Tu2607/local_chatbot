package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"

	"local_chatbot/server/helper"
)

type ChatRequest struct {
	Input     string `json:"input"`
	Model     string `json:"model"`
	SessionID string `json:"sessionID"`
}

type ChatResponse struct {
	Response string `json:"response"`
}

var availableGeminiModels = []string{
	"gemma-3-27b-it",
	"gemini-2.5-flash",
	"gemini-2.5-pro",
	"gemini-2.5-flash-lite",
	"gemini-2.0-flash-preview-image-generation",
}

var availableOpenAIModels = []string{
	"gpt-4o",
	"gpt-4.1-mini",
}

var availableOllamaModels = []string{
	"llama3.2:1b",
}

func ChatHandler(redis_session_manager *RedisSessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method. Only POST is allowed", http.StatusMethodNotAllowed)
			return
		}

		var req ChatRequest
		// Decode the JSON request body, filling out the fields
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Failed to decode request body", http.StatusBadRequest)
			return
		}

		// Check if the response should be in HTML format
		isHTML := r.URL.Query().Get("format") == "html"

		sessionID := req.SessionID
		if sessionID == "" {
			log.Println("Session ID is required. Generating new one.")
			sessionID = helper.GenerateULID()
		}

		var resp ChatResponse
		if slices.Contains(availableGeminiModels, req.Model) {
			reply := GeminiHandler(redis_session_manager, sessionID, req.Input, req.Model, isHTML)
			resp = ChatResponse{Response: reply}
		} else if slices.Contains(availableOpenAIModels, req.Model) {
			reply := OpenAIHandler(req.Input, req.Model, isHTML)
			resp = ChatResponse{Response: reply}
			// Call the OpenAI handler function
		} else if slices.Contains(availableOllamaModels, req.Model) {
			reply := OllamaHandler(redis_session_manager, sessionID, req.Input, req.Model, isHTML)
			resp = ChatResponse{Response: reply}
		} else {
			http.Error(w, "Unsupported model", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
