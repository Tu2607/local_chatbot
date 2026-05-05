# Architecture

## Overview

This application is a Go HTTP server for a local chatbot UI. It serves the static frontend from `static/`, exposes chat/session APIs from `server/handler`, stores session state in Redis, and routes model calls through a provider registry that currently supports Gemini and Ollama.

OpenAI is not wired into this app.

## Main Components

- `main.go` is the process entrypoint. It initializes logging, loads configuration, builds the application dependencies, mounts routes, optionally starts a local Ollama server, and manages graceful shutdown.
- `internal/config` loads environment-based configuration. Important env vars include `PORT`, `REDIS_ADDR`, `REDIS_PORT`, `GEMINI_API_KEY`, `DATA_DIR`, and `OLLAMA_HOST`.
- `internal/app` is the dependency container. It owns the config, Redis client, Redis session manager, provider registry, and the wait group used for background context-sync work.
- `internal/provider` defines the provider interface and registry. Providers implement chat calls, supported model discovery, model selection, history compression, provider naming, and cleanup.
- `server/handler` contains HTTP handlers, provider adapters, and Redis-backed session management.
- `server/template` contains request/response DTOs and chat message types shared across handlers and providers.
- `server/utility` contains shared helpers, including the global structured logger, ULID generation, response formatting, slice helpers, and UTF-8 boundary utilities.
- `internal/rag` contains RAG infrastructure: document parsing, ChromaDB storage/search, and RAG-specific types.

## Startup Flow

1. `main.go` initializes `utility.Logger` before any other startup work.
2. `config.Load` reads environment variables and validates basic port settings.
3. `app.New` creates a Redis client and verifies connectivity with `PING`.
4. `app.New` creates a `RedisSessionManager` around that Redis client.
5. `app.New` creates a provider registry.
6. If `GEMINI_API_KEY` is present, `app.New` initializes and registers the Gemini provider.
7. `app.New` always initializes and registers the Ollama provider.
8. `main.go` creates an `http.ServeMux`.
9. The mux serves `static/` at `/`.
10. The mux mounts `POST /chat` and `GET`/`DELETE /session`.
11. If `OLLAMA_HOST` resolves to `localhost`, `main.go` attempts to start `ollama serve`.
12. The HTTP server starts on `PORT`.

## Chat Request Flow

The main runtime path is `POST /chat`.

1. The frontend sends a JSON `ChatRequest` with `input`, `model`, and optional `sessionID`.
2. `ChatHandler` rejects non-POST requests and decodes the request body.
3. If `sessionID` is absent, the handler generates a new ULID.
4. If `model` is absent, the handler tries to load the previously selected model from Redis field `currmodel`.
5. The provider registry resolves the provider by checking which provider supports the selected model.
6. The handler calls `SetModel` on the selected provider.
7. The selected model is saved back to Redis under the session hash field `currmodel`.
8. The existing conversation history is loaded from Redis hash field `history`.
9. The handler calls `Provider.SendMessage` with the session ID, user input, previous history, and response format preference.
10. The provider converts the shared `template.Message` history into provider-specific format and calls Gemini or Ollama.
11. The handler appends the user message and model response to the in-memory history.
12. A background goroutine compresses long history through the same provider and saves the compressed result to Redis.
13. The handler immediately returns JSON `{ "response": "..." }` without waiting for compression.

## Session Flow

The `/session` endpoint exposes stored Redis session state.

- `GET /session?key=allid` scans Redis keys, sorts them, reverses the result, and returns the known session IDs.
- `GET /session?key={sessionID}` loads the Redis `history` and `currmodel` fields for that session and returns them as JSON.
- `DELETE /session?key={sessionID}` deletes the Redis key for that session.

Redis stores each session as a hash keyed by session ID. Known fields are:

- `history`: JSON array of `template.Message`
- `currmodel`: selected model for the session

## Provider Flow

Providers are hidden behind `internal/provider.Provider`.

The registry does not route by provider name during chat. It routes by model:

1. `Registry.GetByModel(model)` loops over registered providers.
2. Each provider exposes `GetSupportedModels`.
3. The first provider that supports the requested model handles the request.

Gemini is registered only when `GEMINI_API_KEY` is configured. Ollama is always registered, using the Ollama client environment configuration.

## RAG Flow

RAG support exists as infrastructure, but it is not yet part of the `/chat` request path.

The intended storage boundary is session-scoped:

- Uploaded document chunks live in ChromaDB collection `session:{sessionID}:documents`.
- Archived chat lives in ChromaDB collection `session:{sessionID}:chat`.
- Callers pass a `sessionID`; collection naming stays centralized in `internal/rag/chromadb_client.go`.

Current RAG pieces:

- `internal/rag/parser.go` parses TXT, PDF, and DOCX files into byte-based chunks.
- PDF parsing shells out to `pdftotext`.
- DOCX parsing reads `word/document.xml` from the DOCX zip and extracts WordprocessingML text nodes.
- Chunking uses UTF-8 boundary helpers from `server/utility`.
- `ChromaDBClient` stores and searches document chunks and archived chat using ChromaDB built-in embeddings.

Future RAG injection should preserve this context priority:

1. Recent chat from Redis
2. Archived chat from ChromaDB
3. Uploaded document chunks from ChromaDB

## Shutdown Flow

1. `main.go` waits for `SIGINT` or `SIGTERM`.
2. It creates a five-second shutdown context.
3. It waits for background context-sync work, such as history compression, through `App.WaitForContextSync`.
4. It gracefully shuts down the HTTP server.
5. Deferred cleanup closes the provider registry, closes Redis, and stops the local Ollama process if this process started one.
