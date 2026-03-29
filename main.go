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

	"local_chatbot/internal/app"
	"local_chatbot/internal/config"
	"local_chatbot/server/handler"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize application
	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer application.Close()

	// Setup HTTP routes
	mux := http.NewServeMux()

	// Serve static files
	mux.Handle("/", http.FileServer(http.Dir("./static/")))

	// Setup API endpoints with new handler that uses the provider registry
	mux.HandleFunc("/chat", handler.ChatHandler(application.SessionManager, application.ProviderRegistry, &application.WgContextSync))
	mux.HandleFunc("/session", handler.SessionHandler(application.SessionManager))

	// Start Ollama server (optional)
	ollamaCmd := startOllamaServer()
	defer stopOllamaServer(ollamaCmd)

	// Create HTTP server
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	// Signal handling for graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		log.Println("Starting server on port " + cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-sig

	// Graceful shutdown
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := application.WaitForContextSync(ctx); err != nil {
		log.Printf("Warning: Context sync did not complete before shutdown: %v", err)
	}

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Error shutting down server: %v", err)
	}
	log.Println("Server gracefully stopped")
}

// Helper function to start Ollama server
func startOllamaServer() *os.Process {
	ollamaCmd := exec.Command("ollama", "serve")
	if err := ollamaCmd.Start(); err != nil {
		log.Println("Warning: Ollama server could not be started. Ollama may not be installed or accessible in PATH.")
		return nil
	}
	log.Println("Ollama server started")
	return ollamaCmd.Process
}

// Helper function to stop Ollama server
func stopOllamaServer(process *os.Process) {
	if process != nil {
		if err := process.Kill(); err != nil {
			log.Printf("Warning: Error stopping Ollama server: %v", err)
		} else {
			log.Println("Ollama server stopped")
		}
	}
}
