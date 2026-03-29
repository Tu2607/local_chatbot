package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"local_chatbot/internal/provider"
	"local_chatbot/server/helper"
	"local_chatbot/server/template"
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
			log.Println("Session ID is required. Generating new one.")
			sessionID = helper.GenerateULID()
		}

		// If model is not provided in the request, try to fetch it from the session.
		if req.Model == "" {
			savedModel, err := sessionManager.GetSessionModel(ctx, sessionID)
			if err != nil {
				log.Printf("Error: %v", err)
			} else {
				req.Model = savedModel
				log.Printf("No model provided in request, using saved model from session: %s", req.Model)
			}
		}

		// Get the provider for the requested model
		prov, err := providerRegistry.GetByModel(req.Model)
		if err != nil {
			http.Error(w, "Unsupported model: "+req.Model, http.StatusBadRequest)
			return
		}

		// Set the selected model for the provider (if applicable)
		if err := prov.SetModel(req.Model); err != nil {
			log.Printf("Error setting model for provider: %v", err)
			http.Error(w, "Error setting model for provider", http.StatusInternalServerError)
			return
		}

		// Save model to session
		if err := sessionManager.SaveSessionModel(ctx, sessionID, req.Model); err != nil {
			log.Printf("Error saving session model: %v", err)
		}

		// Fetch session history
		history, err := sessionManager.GetSessionHistory(ctx, sessionID)
		if err != nil {
			log.Printf("Error fetching session history: %v", err)
			http.Error(w, "Error fetching session history", http.StatusInternalServerError)
			return
		}

		// Send message to provider
		reply, err := prov.SendMessage(ctx, sessionID, req.Input, history, isHTML)
		if err != nil {
			log.Printf("Error sending message to provider: %v", err)
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
				log.Printf("Error compressing session history: %v", err)
			}
			if err := sessionManager.SaveSessionHistory(ctx, sessionID, "history", compressedHistory); err != nil {
				log.Printf("Error saving compressed session history: %v", err)
			}
		}()

		// Return response
		resp := template.ChatResponse{Response: reply}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
