package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"local_chatbot/internal/provider"
	"local_chatbot/server/template"
	"local_chatbot/server/utility"
)

func ChatHandler(sessionManager *RedisSessionManager, providerRegistry *provider.Registry, contextSyncWg *sync.WaitGroup) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method. Only POST is allowed", http.StatusMethodNotAllowed)
			return
		}

		var req template.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Failed to decode request body", http.StatusBadRequest)
			return
		}

		// Check if the response should be in HTML format
		isHTML := r.URL.Query().Get("format") == "html"

		sessionID := req.SessionID
		if sessionID == "" {
			utility.Logger.WithComponent("chat_handler").Info("No session ID provided in request, generating new session ID")
			sessionID = utility.GenerateULID()
			utility.Logger.WithComponent("chat_handler").WithSessionID(sessionID).Debug("Generated new session ID")
		}

		// If model is not provided in the request, try to fetch it from the session.
		if req.Model == "" {
			savedModel, err := sessionManager.GetSessionModel(ctx, sessionID)
			if err != nil {
				utility.Logger.WithComponent("chat_handler").Error(err, "Error fetching session model")
			} else {
				req.Model = savedModel
				utility.Logger.WithComponent("chat_handler").Info("No model provided in request, using saved model from session", "model", req.Model)
			}
		}

		// Get the provider for the requested model
		prov, err := providerRegistry.GetByModel(req.Model)
		if err != nil {
			http.Error(w, "Unsupported model: "+req.Model, http.StatusBadRequest)
			return
		}
		utility.Logger.WithComponent("chat_handler").WithSessionID(sessionID).Debug("Using provider for model", "model", req.Model)

		// Set the selected model for the provider (if applicable)
		if err := prov.SetModel(req.Model); err != nil {
			utility.Logger.WithComponent("chat_handler").Error(err, "Error setting model for provider")
			http.Error(w, "Error setting model for provider", http.StatusInternalServerError)
			return
		}

		// Save model to session
		if err := sessionManager.SaveSessionModel(ctx, sessionID, req.Model); err != nil {
			utility.Logger.WithComponent("chat_handler").Error(err, "Error saving session model", "model", req.Model)
		}

		// Fetch session history
		history, err := sessionManager.GetSessionHistory(ctx, sessionID)
		if err != nil {
			utility.Logger.WithComponent("chat_handler").Error(err, "Error fetching session history")
			http.Error(w, "Error fetching session history", http.StatusInternalServerError)
			return
		}

		// Send message to provider
		reply, err := prov.SendMessage(ctx, sessionID, req.Input, history, isHTML)
		if err != nil {
			utility.Logger.WithComponent("chat_handler").Error(err, "Error sending message to provider")
			http.Error(w, "Error processing request", http.StatusInternalServerError)
			return
		}

		// Save updated history to session
		history = append(history, template.Message{Content: req.Input, Role: "user"})
		history = append(history, template.Message{Content: reply, Role: "model"})

		// Compress history for long conversations to save tokens, this will be used in the next conversation turn
		// Run asynchronously since we don't need to wait for this to return before sending the response back to the user
		contextSyncWg.Add(1)
		go func() {
			defer contextSyncWg.Done()
			compressedHistory, err := prov.CompressHistory(history)
			if err != nil {
				utility.Logger.WithComponent("chat_handler").WithSessionID(sessionID).Error(err, "Error compressing session history")
			}
			if err := sessionManager.SaveSessionHistory(ctx, sessionID, "history", compressedHistory); err != nil {
				utility.Logger.WithComponent("chat_handler").WithSessionID(sessionID).Error(err, "Error saving compressed session history")
			}
		}()

		// Return response
		resp := template.ChatResponse{Response: reply}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
