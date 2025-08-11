package handler

import (
	"context"
	"encoding/json"
	"local_chatbot/server/template"
	"log"

	"github.com/redis/go-redis/v9"
)

type RedisSessionManager struct {
	client *redis.Client
}

// Initialize a new Redis session manager with the provided Redis client
func NewRedisSessionManager(client *redis.Client) *RedisSessionManager {
	return &RedisSessionManager{
		client: client,
	}
}

func (rsm *RedisSessionManager) GetSessionHistory(ctx context.Context, sessionID string) ([]*template.Message, error) {
	historyData, err := rsm.client.Get(ctx, sessionID).Result()
	if err == redis.Nil {
		return []*template.Message{}, nil // No history found for the session, return an empty slice of template.Message
	} else if err != nil {
		return nil, err // Return the error if there was an issue retrieving the session history
	}

	// Deserialize the history data from JSON format
	var history []*template.Message
	if err := json.Unmarshal([]byte(historyData), &history); err != nil {
		return nil, err // Return the error if deserialization fails
	}

	return history, nil
}

func (rsm *RedisSessionManager) SaveSessionHistory(ctx context.Context, sessionID string, history []*template.Message) error {
	historyData, err := json.Marshal(history)
	if err != nil {
		log.Fatal("Failed to marshal session history:", err)
	}

	// Store the updated session history in Redis
	return rsm.client.Set(ctx, sessionID, historyData, 0).Err()
}
