package handler

import (
	"context"
	"encoding/json"
	"log"
	"sort"

	"local_chatbot/server/helper"
	"local_chatbot/server/template"

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

func (rsm *RedisSessionManager) GetSessionHistory(ctx context.Context, sessionID string, key string) ([]template.Message, error) {
	historyData, err := rsm.client.HGet(ctx, sessionID, key).Result()
	if err == redis.Nil {
		return []template.Message{}, nil // No history found for the session, return an empty slice of template.Message
	} else if err != nil {
		return nil, err // Return the error if there was an issue retrieving the session history
	}

	// Deserialize the history data from JSON format
	var history []template.Message
	if err := json.Unmarshal([]byte(historyData), &history); err != nil {
		return nil, err // Return the error if deserialization fails
	}

	return history, nil
}

// Scan the redis cache and returns all keys in a sorted manner with the most recent being first
func (rsm *RedisSessionManager) GetAllSessionsID(ctx context.Context, pattern string) ([]string, error) {
	var cursor uint64
	var allKeys []string

	for {
		var err error
		keys, nextCursor, err := rsm.client.Scan(ctx, cursor, pattern, 1000).Result()
		if err != nil {
			log.Fatal("Failed to scan Redis keys:", err)
		}

		allKeys = append(allKeys, keys...)

		if nextCursor == 0 {
			log.Println("Successfully scanned all Redis keys")
			sort.Strings(allKeys) // Sort the keys to have the most recent first
			return helper.ReverseSlice(allKeys), nil
		}
		cursor = nextCursor
	}

}

func (rsm *RedisSessionManager) DeleteSession(ctx context.Context, pattern string) error {
	// Delete the session from Redis
	return rsm.client.Del(ctx, pattern).Err()
}

func (rsm *RedisSessionManager) SaveSessionHistory(ctx context.Context, sessionID string, key string, history []template.Message) error {
	historyData, err := json.Marshal(history)
	if err != nil {
		log.Fatal("Failed to marshal session history:", err)
	}

	return rsm.client.HSet(ctx, sessionID, key, historyData).Err()
}

func (rsm *RedisSessionManager) SaveSessionModel(ctx context.Context, sessionID string, model string) error {
	return rsm.client.HSet(ctx, sessionID, "model", model).Err()
}

func (rsm *RedisSessionManager) GetSessionModel(ctx context.Context, sessionID string) (string, error) {
	modelData, err := rsm.client.HGet(ctx, sessionID, "model").Result()
	if err == redis.Nil {
		return "", nil // No model found for the session
	} else if err != nil {
		return "", err
	}

	return modelData, nil
}
