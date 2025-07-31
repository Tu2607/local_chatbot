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

	port := "8080"
	fmt.Println("Starting server on port", port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}
