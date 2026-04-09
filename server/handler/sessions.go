package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"local_chatbot/server/template"
	"local_chatbot/server/utility"
)

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
					utility.Logger.WithComponent("session_handler").Error(err, "Failed to get all session IDs")
					http.Error(w, "Failed to get all sessions", http.StatusInternalServerError)
					return
				}

				resp := template.AllSessionRequest{Keys: sessionsID}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			} else if key != "" {
				// Handle GET requests for a specific session chat context
				ctx := context.Background()
				history, err := redisSessionManager.GetSessionHistory(ctx, key)

				if err != nil {
					http.Error(w, "Failed to get session history", http.StatusInternalServerError)
					return
				}

				model, err := redisSessionManager.GetSessionModel(ctx, key)

				if err != nil {
					http.Error(w, "Failed to get session model", http.StatusInternalServerError)
					return
				}

				// If the session history is empty, return a 404 response
				if len(history) == 0 {
					http.Error(w, "Session history not found", http.StatusNotFound)
					return
				}

				// Return the session history
				resp := &template.SessionContextRequest{ChatHistory: history, Model: model}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			} else {
				http.Error(w, "Invalid session key", http.StatusBadRequest)
			}
		case http.MethodDelete:
			// Handle DELETE requests for a specific session
			key := r.URL.Query().Get("key")
			if key == "" {
				utility.Logger.WithComponent("session_handler").Warn("No session key provided for DELETE request")
				http.Error(w, "Missing session key", http.StatusBadRequest)
				return
			}

			ctx := context.Background()
			if err := redisSessionManager.DeleteSession(ctx, key); err != nil {
				http.Error(w, "Failed to delete session", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusNoContent)
		}
	}
}
