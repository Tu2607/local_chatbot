# Local AI Chat Bot

Hi there! This is the repository for the local chat bot that you can locally host on your network. In its current iteration of development, this chat bot can be a central hub where users can use the Gemini or OpenAI API or the yet to be implemented local model through Ollama. 


# Instructions And Requirements

To run this chat bot, you will need to install Go, as the back-end is written completely in Go. Instruction on how to install Go on your machine can be found at https://go.dev/doc/install.

Before compiling the code, please ensure that you have the API keys from services like Gemini or OpenAI set in your machine's environment variables. For Gemini API, the env-var that the back-end code will be looking for is **GEMINI_API_KEY**. For OpenAI API, the env-var that the back-end code will be looking for is **OPENAI_API_KEY**.

If you are looking to use a local model through Ollama, then please ensure that the Ollama service is running on the machine before starting server. At this point, support for using local model through Ollama is not yet implement (sorry!).
 
With Go installed on your machine, you can compile the code from the repository main directory by doing:
> go build main.go

Then run the server by doing:
> ./main

Please ensure that port 55572 is open on your machine so that the chat bot can be reachable from other machines in your network. You can also edit what you want the server to listen to by modifying in the **main.go** file.

## Current Supported Models

The Chat Bot currently supports most of the models from the free tier of Gemini API:
- Gemini 2.5 Pro
- Gemini 2.5 Flash
- Gemini 2.5 Flash Lite
- Gemma 3 (27B)
- Gemini 2.0 Flash Preview Image Generation

Currently in the works for OpenAI API:
- GPT-4o
- GPT-3.5



## Contributions

Any contributions is greatly appreciated as this project is a learning project to learn Go and the inner working of a web server.
