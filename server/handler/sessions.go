package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"local_chatbot/server/helper"
	"local_chatbot/server/template"
)

type SessionContextRequest struct {
	ChatHistory []template.Message `json:"context"`
}

type AllSessionRequest struct {
	Keys []string `json:"keys"`
}

func SessionHandler(redisSessionManager *RedisSessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handle session-related requests
		if r.Method != http.MethodGet && r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed, only GET and DELETE are allowed", http.StatusMethodNotAllowed)
			return
		}

		switch r.Method {
		case http.MethodGet:
			// Each session is represent as a key in the redis cache
			key := r.URL.Query().Get("key")
			if key == "allid" {
				// Handle GET requests to list all sessions ID
				sessionsID, err := redisSessionManager.GetAllSessionsID(context.Background(), "*")
				if err != nil {
					http.Error(w, "Failed to get all sessions", http.StatusInternalServerError)
					return
				}

				resp := AllSessionRequest{Keys: sessionsID}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			} else if key != "" {
				// Handle GET requests for a specific session chat context
				ctx := context.Background()
				history, err := redisSessionManager.GetSessionHistory(ctx, key, "complete")
				if err != nil {
					http.Error(w, "Failed to get session history", http.StatusInternalServerError)
					return
				}

				// If the session history is empty, return a 404 response
				if len(history) == 0 {
					http.Error(w, "Session history not found", http.StatusNotFound)
					return
				}

				// Return the session history
				resp := SessionContextRequest{ChatHistory: history}

				// Because we can't modify `resp` in place since it's not a reference,
				// we need to create a new response object with the modified content that parsed the text to HTML.
				// If the request has a query parameter `format=html`, we will convert the content to HTML
				if isHTML := r.URL.Query().Get("format") == "html"; isHTML {
					htmlResp := make([]template.Message, len(resp.ChatHistory))
					for i, msg := range resp.ChatHistory {
						htmlContent := helper.HtmlOrCurlResponse(isHTML, msg.Content)
						htmlResp[i] = template.Message{Role: msg.Role, Content: htmlContent}
					}
					resp = SessionContextRequest{ChatHistory: htmlResp}
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			} else {
				http.Error(w, "Invalid session key", http.StatusBadRequest)
			}
		case http.MethodDelete:
			// Handle DELETE requests
		}
	}
}
