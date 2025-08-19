package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"local_chatbot/server/template"
)

type SessionRequest struct {
	Key []template.Message `json:"key"`
}

type AllSessionRequest struct {
	Keys []string `json:"keys"`
}

type AllSessionContextRequest struct {
	Contexts []template.Message `json:"contexts"`
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
				// } else if key == "allcontext" {
				// allSessionID, err := redisSessionManager.GetAllSessionsID(context.Background(), "*")
				// if err != nil {
				// http.Error(w, "Failed to get all sessions IDs", http.StatusInternalServerError)
				// return
				// }

				// // Now that we have all session IDs, we can retrieve their histories/chat contexts
				// var allHistories []template.Message
				// for _, sessionID := range allSessionID {
				// history, err := redisSessionManager.GetSessionHistory(context.Background(), sessionID)
				// if err != nil {
				// http.Error(w, "Failed to get session history", http.StatusInternalServerError)
				// return
				// }
				// allHistories = append(allHistories, history...)
				// }
				// resp := AllSessionContextRequest{Contexts: allHistories}

				// w.Header().Set("Content-Type", "application/json")
				// json.NewEncoder(w).Encode(resp)
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
				resp := SessionRequest{Key: history}
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
