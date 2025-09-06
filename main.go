package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"local_chatbot/server/handler"

	"github.com/redis/go-redis/v9"
)

func main() {
	// Signal handling for graceful shutdown
	// ...
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

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

	// Check if Ollama is installed by doing ollama serve
	ollamaCmd := exec.Command("ollama", "serve")
	if err := ollamaCmd.Start(); err != nil {
		log.Println("Error starting Ollama server. Ollama is not installed or not found in PATH. Please install Ollama to use Ollama models.")
	} else {
		log.Println("Ollama server is running.")
	}

	port := "55572"
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Handle the APIs
	mux.HandleFunc("/chat", handler.ChatHandler(redis_session_manager))
	mux.HandleFunc("/session", handler.SessionHandler(redis_session_manager))

	go func() {
		log.Println("Starting server on port " + port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-sig
	// Graceful shutdown
	log.Println("Shutting down server...")
	// Shut down the http server, close database connections, etc.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Error on server shutting down: %v", err)
	}
	log.Println("Server gracefully stopped")

	// Close Redis connection
	log.Printf("Closing Redis connection...")
	if err := redis_db.Close(); err != nil {
		log.Fatalf("Error closing Redis connection: %v", err)
	}
	log.Println("Redis connection closed.")

	// Shutdown Ollama server if it was started
	if ollamaCmd.Process != nil {
		if err := ollamaCmd.Process.Kill(); err != nil {
			log.Printf("Error stopping Ollama server: %v", err)
		} else {
			log.Println("Ollama server stopped.")
		}
	}
}
