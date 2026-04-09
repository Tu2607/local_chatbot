package main

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"local_chatbot/internal/app"
	"local_chatbot/internal/config"
	"local_chatbot/server/handler"
	"local_chatbot/server/utility"
)

func main() {
	// Initialize logger
	utility.Logger = utility.InitLogger(os.Getenv("DEBUG") == "true")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		utility.Logger.WithComponent("main").Fatal("Failed to load configuration", "error", err)
	}

	// Initialize application
	application, err := app.New(cfg)
	if err != nil {
		utility.Logger.WithComponent("main").Fatal("Failed to start application", "error", err)
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
	var ollamaCmd *os.Process
	if cfg.GetOllamaHost() == "localhost" {
		ollamaCmd = startOllamaServer()
	} else {
		utility.Logger.WithComponent("main").Info("Ollama host is not localhost, skipping starting local Ollama server", "ollama_host", cfg.GetOllamaHost())
	}
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
		utility.Logger.WithComponent("main").Info("Starting server", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utility.Logger.WithComponent("main").Fatal("Failed to start server", "error", err)
		}
	}()

	// Wait for shutdown signal
	<-sig

	// Graceful shutdown
	utility.Logger.WithComponent("main").Info("Shutdown signal received, shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := application.WaitForContextSync(ctx); err != nil {
		utility.Logger.WithComponent("main").Warn("Context sync did not complete before shutdown", "error", err)
	}

	if err := server.Shutdown(ctx); err != nil {
		utility.Logger.WithComponent("main").Fatal("Error shutting down server", "error", err)
	}
	utility.Logger.WithComponent("main").Info("Server gracefully stopped")
}

// Helper function to start Ollama server
func startOllamaServer() *os.Process {
	ollamaCmd := exec.Command("ollama", "serve")
	if err := ollamaCmd.Start(); err != nil {
		utility.Logger.WithComponent("main").Warn("Failed to start Ollama server", "error", err)
		return nil
	}
	utility.Logger.WithComponent("main").Info("Ollama server started")
	return ollamaCmd.Process
}

// Helper function to stop Ollama server
func stopOllamaServer(process *os.Process) {
	if process != nil {
		if err := process.Kill(); err != nil {
			utility.Logger.WithComponent("main").Warn("Error stopping Ollama server", "error", err)
		} else {
			utility.Logger.WithComponent("main").Info("Ollama server stopped")
		}
	}
}
