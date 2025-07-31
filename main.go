package main

import (
	"fmt"
	"local_chatbot/server/handler"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	file_server := http.FileServer(http.Dir("./static/"))

	// Serve static files from the "static" directory
	mux.Handle("/", file_server)

	// Handle the chat api
	mux.HandleFunc("/chat", handler.ChatHandler)

	port := "55572"
	fmt.Println("Starting server on port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
