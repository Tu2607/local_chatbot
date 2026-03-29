package template

// This file defines the data structures used for communication between the client and server,
// as well as between different components of the server.

type Message struct {
	Content string `json:"content"`
	Role    string `json:"role"` // "user" or "assistant"
}

type ChatRequest struct {
	Input     string `json:"input"`
	Model     string `json:"model"`
	SessionID string `json:"sessionID"`
}

type ChatResponse struct {
	Response string `json:"response"`
}

type SessionContextRequest struct {
	ChatHistory []Message `json:"context"`
	Model       string    `json:"model"`
}

type AllSessionRequest struct {
	Keys []string `json:"keys"`
}
