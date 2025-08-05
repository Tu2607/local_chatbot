package handler

import (
	"encoding/json"
	"local_chatbot/server/ai_models"
	"local_chatbot/server/helper"
	"net/http"

	"github.com/google/uuid"
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
	// Decode the JSON request body, filling out the fields
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	// Check if cookie for session ID exists, if not create a new one
	cookie, err := r.Cookie("sessionID")
	var sessionID string

	if err != nil || cookie.Value == "" {
		sessionID = uuid.New().String()
		http.SetCookie(w, &http.Cookie{
			Name:     "sessionID",
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
		})
	} else {
		sessionID = cookie.Value
	}

	sessions.RLock()
	history := sessions.histories[sessionID]
	sessions.RUnlock()

	// Check if the response should be in HTML format
	isHTML := r.URL.Query().Get("format") == "html"

	// Initialize the response variables
	var resp ChatResponse

	switch req.Model {
	case "gemma-3-27b-it":
		reply, updatedHistory := ai_models.GeminiChat(history, req.Input, req.Model)
		sessions.Lock()
		sessions.histories[sessionID] = updatedHistory // Update the session history atomically
		sessions.Unlock()
		resp = ChatResponse{Response: helper.HtmlOrCurlResponse(isHTML, reply)}
	case "gemini-2.5-flash":
		reply, updatedHistory := ai_models.GeminiChat(history, req.Input, req.Model)
		sessions.Lock()
		sessions.histories[sessionID] = updatedHistory // Update the session history atomically
		sessions.Unlock()
		resp = ChatResponse{Response: helper.HtmlOrCurlResponse(isHTML, reply)}
	case "gemini-2.5-pro":
		reply, updatedHistory := ai_models.GeminiChat(history, req.Input, req.Model)
		sessions.Lock()
		sessions.histories[sessionID] = updatedHistory // Update the session history atomically
		sessions.Unlock()
		resp = ChatResponse{Response: helper.HtmlOrCurlResponse(isHTML, reply)}
	case "gemini-2.5-flash-lite": // The cheapest model
		reply, updatedHistory := ai_models.GeminiChat(history, req.Input, req.Model)
		sessions.Lock()
		sessions.histories[sessionID] = updatedHistory // Update the session history atomically
		sessions.Unlock()
		resp = ChatResponse{Response: helper.HtmlOrCurlResponse(isHTML, reply)}
	case "gemini-2.0-flash-preview-image-generation":
		// No need to update history for image generation, I'm assuming the context is not needed
		// if the prompt is sufficiently descriptive enough.
		// Well, maybe in the future when I have time to implement it.
		reply := ai_models.GeminiImageGeneration(req.Input, req.Model)
		encodedReply := helper.EncodeByteSliceToBase64(reply)
		resp = ChatResponse{Response: encodedReply}
	case "local":
		// reply = ai_models.OllamaChat(req.Input, "")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
