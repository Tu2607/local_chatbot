package handler

import (
	"sync"

	"google.golang.org/genai"
)

// Creating a struct to hold information of a chat session
var sessions = struct {
	sync.RWMutex
	histories map[string][]*genai.Content // The key is the session ID, and the value is a slice of Content representing the chat history
}{
	histories: make(map[string][]*genai.Content),
}
