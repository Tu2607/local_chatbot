package handler

import (
	"encoding/json"
	"net/http"
	"local_chatbot/models"
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

	// TODO: Implement chat logic here
	switch req.Model {
	case "gemini-2.5-flash":
		models.geminiChat(req.Input, "gemini-2.5-flash")
	case "gemini-2.5-pro":
		models.geminiChat(req.Input, "gemini-2.5-pro")
	case "local:
	}

	resp := ChatResponse{Response: reply}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
