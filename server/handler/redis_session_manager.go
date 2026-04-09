package handler

import (
	"context"
	"encoding/json"
	"sort"

	"local_chatbot/server/template"
	"local_chatbot/server/utility"

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

func (rsm *RedisSessionManager) GetSessionHistory(ctx context.Context, sessionID string) ([]template.Message, error) {
	historyData, err := rsm.client.HGet(ctx, sessionID, "history").Result()
	if err == redis.Nil {
		utility.Logger.WithSessionID(sessionID).Debug("No history found for session")
		return []template.Message{}, nil // No history found for the session, return an empty slice of template.Message
	} else if err != nil {
		return nil, err // Return the error if there was an issue retrieving the session history
	}

	// Deserialize the history data from JSON format
	var history []template.Message
	if err := json.Unmarshal([]byte(historyData), &history); err != nil {
		return nil, err
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
			utility.Logger.WithComponent("redis_session_manager").Error(err, "Failed to scan Redis keys", "cursor", cursor)
			return nil, err
		}

		allKeys = append(allKeys, keys...)

		if nextCursor == 0 {
			utility.Logger.WithComponent("redis_session_manager").Info("Completed scanning Redis keys", "total_keys", len(allKeys))
			sort.Strings(allKeys) // Sort the keys to have the most recent first
			return utility.ReverseSlice(allKeys), nil
		}
		cursor = nextCursor
	}

}

func (rsm *RedisSessionManager) DeleteSession(ctx context.Context, pattern string) error {
	// Delete the session from Redis
	err := rsm.client.Del(ctx, pattern).Err()
	if err != nil {
		utility.Logger.WithComponent("redis_session_manager").Error(err, "Failed to delete session from Redis", "pattern", pattern)
		return err
	}
	return nil
}

// Save the session history to Redis. Key is the session ID and value is the history is the slice of template.Message
func (rsm *RedisSessionManager) SaveSessionHistory(ctx context.Context, sessionID string, key string, history []template.Message) error {
	historyData, err := json.Marshal(history)
	if err != nil {
		utility.Logger.WithSessionID(sessionID).Error(err, "Failed to marshal session history", "key", key)
		return err
	}

	err = rsm.client.HSet(ctx, sessionID, key, historyData).Err()
	if err != nil {
		utility.Logger.WithSessionID(sessionID).Error(err, "Failed to save session history to Redis", "key", key)
		return err
	}

	return nil
}

func (rsm *RedisSessionManager) SaveSessionModel(ctx context.Context, sessionID string, model string) error {
	err := rsm.client.HSet(ctx, sessionID, "currmodel", model).Err()
	if err != nil {
		utility.Logger.WithSessionID(sessionID).Error(err, "Failed to save session model", "model", model)
		return err
	}
	return nil
}

func (rsm *RedisSessionManager) GetSessionModel(ctx context.Context, sessionID string) (string, error) {
	modelData, err := rsm.client.HGet(ctx, sessionID, "currmodel").Result()
	if err == redis.Nil {
		utility.Logger.WithSessionID(sessionID).Debug("No model found for session")
		return "", nil // No model found for the session
	} else if err != nil {
		utility.Logger.WithSessionID(sessionID).Error(err, "Failed to retrieve session model")
		return "", err
	}

	return modelData, nil
}
