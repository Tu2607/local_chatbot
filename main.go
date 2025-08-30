package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"local_chatbot/server/handler"

	"github.com/redis/go-redis/v9"
)

func main() {
	// Signal handling for graceful shutdown
	// ...
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
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

		// Check if Ollama is installed
		cmd := exec.Command("ollama", "-v")
		err := cmd.Run()
		if err != nil {
			log.Printf("Ollama is not installed: %v", err)
		} else {
			// Start the local Ollama server if needed
			// Simply run command `ollama serve`
			cmd := exec.Command("ollama", "serve")
			err := cmd.Start()
			if err != nil {
				log.Printf("Failed to start Ollama server: %v", err)
			} else {
				log.Println("Ollama server started successfully")
			}
		}

		// Handle the APIs
		mux.HandleFunc("/chat", handler.ChatHandler(redis_session_manager))
		mux.HandleFunc("/session", handler.SessionHandler(redis_session_manager))

		port := "55572"
		fmt.Println("Starting server on port", port)
		if err := http.ListenAndServe(":"+port, mux); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-sig
	// Graceful shutdown
	log.Println("Shutting down server...")
	// Shut down the ollama server
	cmd := exec.Command("pkill", "-f", "ollama")
	err := cmd.Run()
	if err != nil {
		log.Printf("Failed to shut down Ollama server: %v", err)
	} else {
		log.Println("Ollama server shut down successfully")
	}
}
