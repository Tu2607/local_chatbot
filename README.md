# Local Chatbot

A multi-provider local AI chatbot application built in Go, supporting Gemini, OpenAI, and Ollama models. Features comprehensive logging, session management with Redis, and RAG (Retrieval-Augmented Generation) capabilities with ChromaDB for document-aware conversations.

## Quick Start

### Requirements

- **Go 1.24+** ([install](https://go.dev/doc/install))
- **Docker & Docker Compose** (for Redis, ChromaDB, Ollama)
- **API Keys** (if using Gemini)

### Environment Setup

1. **Set API Keys:**
   ```bash
   export GEMINI_API_KEY="your-gemini-key"
   export DEBUG=true  # Optional: enables verbose logging
   ```

2. **Start Services (Docker Compose):**
   ```bash
   docker-compose up -d
   ```
   This starts:
   - Redis on 6379 (persistent session storage)
   - ChromaDB on 8000 (vector storage for RAG)
   - Ollama on 11434 (local models)

3. **Build & Run:**
   ```bash
   go build -o local_chatbot main.go
   ./local_chatbot
   ```

4. **Access the App:**
   Open `http://localhost:55572` in your browser

---

## Features

### ✅ Multi-Provider Chat
- **Gemini API** — Latest models (2.5 Pro, 2.5 Flash, Gemma 3, image generation)
- **Ollama** — Local models (Llama 3.2, etc.)
- **Per-session model persistence** — Remembers your selected model

### ✅ Session Management
- Redis-backed persistent storage
- Session history with automatic compression
- Real-time chat display with MathJax + DOMPurify rendering
- Session deletion with proper cleanup

### ✅ Comprehensive Logging
- Structured logging with slog (Go's standard logger)
- SessionID context throughout the request pipeline
- Component-based log filtering (chat_handler, redis_session_manager, etc.)
- Debug mode via `DEBUG=true` environment variable

### 🔜 RAG (Retrieval-Augmented Generation) - Phase 2.1 (In Progress)
- **Document Upload** — Support for PDF, DOCX, TXT files
- **Intelligent Chunking** — Byte-based chunking with overlap (8KB chunks, 512-byte overlap)
- **Semantic Search** — ChromaDB-powered vector search over documents and chat history
- **Hybrid Storage** — Hot/cold storage (Redis for recent chats, ChromaDB for archives)
- **Archive Modes** — Choose between compression or embeddings for long conversations
- **Session-Scoped** — Documents and archives are isolated per session

---

## Architecture

### Backend Structure

```
local_chatbot/
├── main.go                          # Entry point, server setup
├── internal/
│   ├── config/                      # Configuration management
│   ├── app/                         # Application initialization
│   ├── provider/                    # Provider interface & implementations
│   │   ├── interface.go             # Provider contract
│   │   ├── gemini_provider.go       # Gemini API implementation
│   │   └── ollama_provider.go       # Ollama local model implementation
│   └── rag/                         # RAG infrastructure (Phase 2.1)
│       ├── rag_types.go             # Document, DocumentChunk types
│       ├── chromadb_client.go       # ChromaDB HTTP client wrapper
│       ├── parser.go                # Multi-format document parsing
│       └── document_manager.go      # (Planned) Orchestration layer
├── server/
│   ├── handler/
│   │   ├── chat_handler.go          # Chat request handler
│   │   ├── sessions.go              # Session management endpoints
│   │   ├── redis_session_manager.go # Redis client wrapper
│   │   ├── document_handler.go      # (Planned) Document upload endpoints
│   │   └── rag_context.go           # (Planned) Context injection layer
│   ├── utility/
│   │   ├── logger.go                # Global structured logger
│   │   ├── utils.go                 # Helper functions
│   │   └── (base64, embedding utils)
│   └── template/                    # Communication types
│       └── template.go              # Message, ChatRequest, etc.
├── static/
│   ├── index.html                   # Frontend UI
│   ├── script.js                    # Chat logic & session management
│   ├── style.css                    # Styling
│   └── favicon.ico
└── docker-compose.yaml              # Service orchestration

```

### Data Flow

```
User Input
    ↓
POST /chat (sessionID, model, message)
    ↓
chat_handler.go
  ├─ Load or create session
  ├─ Fetch chat history from Redis
  ├─ Send to selected provider (Gemini/OpenAI/Ollama)
  ├─ Store response in Redis
  ├─ Async: Compress history (or archive to ChromaDB)
  └─ Return response
    ↓
Frontend renders with DOMPurify + MathJax
    ↓
GET /session (retrieve session history for display)
    ↓
Redis lookup → Convert to JSON → Return
```

---

## Current Models

### Gemini API
- Gemini 2.5 Pro
- Gemini 2.5 Flash
- Gemini 2.5 Flash Lite
- Gemma 3 (27B)
- Gemini 2.0 Flash (image generation)

### Ollama (Local)
- Llama 3.2 (1B, 8B, 70B variants)
- Gemma 3 (12B, 27B)
- Qwen-Coder
- And any model available in your Ollama instance

---

## Key Implementation Details

### Session Storage (Redis)
Each session stores:
- `history` — Array of messages (user + bot responses)
- `currmodel` — Currently selected model for the session
- Auto-deleted on user action or after inactivity (configurable)

### Chat History Compression
When history reaches 20 messages:
- **Option A (Compress):** Summarize via provider → Keep in Redis
- **Option B (Archive/Embed):** Store in ChromaDB with embeddings → Delete from Redis
- Selected per-session in settings (planned feature)

### Logging Context
All logs include:
- Timestamp + severity level
- Component name (e.g., "chat_handler", "redis_session_manager")
- SessionID (for request tracing)
- Custom fields (error details, counts, etc.)

Example:
```json
{"time":"2026-04-09T14:30:45Z","level":"INFO","msg":"Message saved to Redis","component":"chat_handler","session_id":"abc123xyz"}
```

---

## Recent Fixes & Enhancements

### ✅ Logger Integration (Phase 1)
- Structured logging with sessionID context
- Removed silent failures in compression/save operations
- Graceful error handling without crashes

### ✅ Frontend Display Fix (Phase 1)
- Fixed model responses not showing in UI
- Removed double HTML conversion in session retrieval
- DOMPurify properly sanitizes backend HTML

### 🔄 RAG Infrastructure (Phase 2.1 - In Progress)
- chromadb_client.go with idiomatic Go patterns (ChromaOperation function type)
- Byte-based document chunking (8KB chunks, 512-byte overlap)
- File size limits: TXT 5MB, PDF/DOCX 10MB
- Session-scoped collections for privacy

---

## Environment Variables

```bash
# API Keys (required for respective providers)
GEMINI_API_KEY=your-key

# Server
DEBUG=true|false          # Enable verbose logging (default: false)
SERVER_PORT=55572         # Port to listen on

# ChromaDB (RAG)
CHROMA_DB_URL=http://localhost:8000

# Ollama
OLLAMA_HOST=0.0.0.0:11434
OLLAMA_MODELS=~/.ollama/models  # Important for macOS
```

---

## Testing

### Manual Testing
1. **Session Creation**: Navigate to UI, send a message
2. **Model Switching**: Select different models, verify persistence
3. **History Loading**: Refresh page, verify history loads correctly
4. **Document Upload** (WIP): Test parsing and embedding

### With Docker
```bash
# View logs
docker-compose logs -f local_chatbot
docker-compose logs -f chromadb
docker-compose logs -f redis

# Rebuild container
docker-compose up -d --build
```

---

## Development Status

### Phase 1 ✅ — Core Chat (COMPLETE)
- Multi-provider support (Gemini, OpenAI, Ollama)
- Session management with Redis
- Structured logging system
- Frontend display fixes

### Phase 2.1 🔄 — RAG Infrastructure (IN PROGRESS)
- [x] Type definitions (Document, DocumentChunk, ChatArchiveResult)
- [x] ChromaDB client wrapper (Add, Search, Delete operations)
- [x] File size validation
- [x] Byte-based chunking framework
- [ ] Multi-format parsers (PDF, DOCX, TXT)
- [ ] Document manager orchestration
- [ ] Backend integration (upload endpoints, context injection)
- [ ] Frontend UI (file upload, document list)

### Phase 2.2 📋 — Backend RAG Integration (PLANNED)
- Provider interface extensions (SendMessageWithContext)
- Document upload/delete endpoints
- Chat context injection with priority handling
- Archive trigger logic in chat_handler

### Phase 2.3 📋 — Frontend RAG UI (PLANNED)
- Document upload component
- Document list management
- RAG toggle checkbox
- End-to-end testing

---

## Contributing

Contributions are welcome! This is a learning project to explore:
- Go web servers and concurrency patterns
- Structured logging and error handling
- LLM integration and prompt engineering
- Vector databases and semantic search
- Frontend/backend communication

---

## Troubleshooting

### Connection Issues
- Verify port 55572 is open: `lsof -i :55572`
- Check Docker services: `docker-compose ps`
- Verify ChromaDB running: `curl http://localhost:8000/api/v1`

### Logging Issues
- Enable debug logs: `export DEBUG=true`
- Check logs: `docker-compose logs local_chatbot`

### API Key Issues
- Verify env vars: `echo $GEMINI_API_KEY`
- Keys must be set **before** starting the app
- Restart server after changing env vars

---

**Last Updated:** April 9, 2026
**Maintainer:** Tuvu
