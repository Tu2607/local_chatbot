package rag

// DocumentChunk represents a chunk of a document for storage in ChromaDB.
// The Metadata field contains info used for retrieval and filtering in ChromaDB.
type DocumentChunk struct {
	Content    []byte    // Chunk text content (or base64-encoded binary for images)
	ID         string    // Unique chunk ID (e.g., "doc_123:chunk_5")
	Embedding  []float32 // Vector embedding from ChromaDB model
	DocumentID string    // Parent document ID for cleanup/tracking
	Metadata   map[string]any
	// Metadata keys when storing in ChromaDB:
	// - "document_id": string (parent document)
	// - "doc_name": string (original filename)
	// - "chunk_index": int (position in document)
	// - "file_type": string (pdf/docx/txt)
	// Optional metadata:
	// - "page": int (if available from parser)
	// - "section": string (if document has sections)
}

// Document represents a complete document before parsing into chunks.
type Document struct {
	ID       string          // Unique document ID
	FileName string          // Original filename
	FileType string          // "pdf", "docx", "txt"
	FileSize int64           // Size in bytes (to filter oversized docs)
	Contents []DocumentChunk // All chunks from parsing
	Metadata map[string]any
	// Metadata keys (optional, for document-level tracking):
	// - "uploaded_at": timestamp string
	// - "session_id": string (though redundant, already in collection name)
}

// ChatArchiveResult represents a single archived chat message retrieved from ChromaDB.
// Used for semantic search over chat history.
type ChatArchiveResult struct {
	Content string // Combined message (e.g., "user: hello" or "assistant: hi there")
	Role    string // Original role ("user" or "assistant")
	Index   int    // Message index in session
}
