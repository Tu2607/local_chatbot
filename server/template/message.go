package template

// Template of a message in the chat history, which can be used for both user and bot messages
type Message struct {
	Content string `json:"content"`
	Role    string `json:"role"` // "user" or "assistant"
}
