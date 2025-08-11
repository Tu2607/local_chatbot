package main

import (
	"fmt"
	"local_chatbot/server/handler"
	"log"
	"net/http"

	"github.com/redis/go-redis/v9"
)

func main() {
	mux := http.NewServeMux()
	file_server := http.FileServer(http.Dir("./static/"))

	// Serve static files from the "static" directory
	mux.Handle("/", file_server)

	// Initialize Redis session manager
	redis_db := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis server address
		Password: "",               // No password set
		DB:       0,                // Use default DB

	})

	redis_session_manager := handler.NewRedisSessionManager(redis_db)

	// Handle the chat api
	mux.HandleFunc("/chat", handler.ChatHandler(redis_session_manager))

	port := "55572"
	fmt.Println("Starting server on port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
