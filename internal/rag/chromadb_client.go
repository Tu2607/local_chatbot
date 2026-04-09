package rag

import (
	"context"
	"fmt"
	"local_chatbot/server/template"
	"local_chatbot/server/utility"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
)

type ChromaDBClient struct {
	Client chroma.Client
}

type ChromaOperation func(collection chroma.Collection) error

// withCollection retrieves or creates a collection and executes the provided operation against it.
// This is an idiomatic Go pattern using function types for composable operations.
func (c *ChromaDBClient) withCollection(collectionName string, op ChromaOperation) error {
	collection, err := c.CreateOrGetCollection(collectionName)
	if err != nil {
		return err
	}
	return op(collection)
}

// getDocumentCollectionName returns the collection name for documents in a session
func (c *ChromaDBClient) getDocumentCollectionName(sessionID string) string {
	return fmt.Sprintf("session:%s:documents", sessionID)
}

// getChatCollectionName returns the collection name for archived chat in a session
func (c *ChromaDBClient) getChatCollectionName(sessionID string) string {
	return fmt.Sprintf("session:%s:chat", sessionID)
}

func NewChromaDBClient(server_address string) (*ChromaDBClient, error) {
	// Initialize a Chroma HTTP client since we're connecting to a ChromaDB server in a container.
	client, err := chroma.NewHTTPClient(
		chroma.WithBaseURL(server_address),
	)
	if err != nil {
		utility.Logger.WithComponent("chromadb_client").Error(err, "Failed to create ChromaDB client")
		return nil, err
	}

	return &ChromaDBClient{
		Client: client,
	}, nil
}

// CreateOrGetCollection creates a new collection in ChromaDB if it doesn't exist, or retrieves the existing collection if it does.
func (c *ChromaDBClient) CreateOrGetCollection(collectionName string) (chroma.Collection, error) {
	collection, err := c.Client.GetOrCreateCollection(context.Background(), collectionName)
	if err != nil {
		utility.Logger.WithComponent("chromadb_client").Error(err, "Failed to create or get collection", "collection_name", collectionName)
		return nil, err
	}
	return collection, nil
}

// AddDocumentChunks adds a single document chunk to a session's document collection.
// Takes sessionID and constructs the collection name internally.
func (c *ChromaDBClient) AddDocumentChunks(sessionID string, chunk DocumentChunk) error {
	collectionName := c.getDocumentCollectionName(sessionID)
	return c.withCollection(collectionName, func(col chroma.Collection) error {
		chunksStr := string(chunk.Content)
		err := col.Add(context.Background(),
			chroma.WithIDs(chroma.DocumentID(chunk.ID)),
			chroma.WithTexts(chunksStr),
			chroma.WithMetadatas(chroma.NewMetadataFromMap(chunk.Metadata)),
		)
		if err != nil {
			utility.Logger.WithComponent("chromadb_client").Error(err, "Failed to add document chunk", "chunk_id", chunk.ID, "session_id", sessionID)
			return err
		}
		return nil
	})
}

// AddDocument adds all chunks of a document to a session's document collection.
func (c *ChromaDBClient) AddDocument(sessionID string, document Document) error {
	// Adding each chunk of the document to the collection
	for _, chunk := range document.Contents {
		if err := c.AddDocumentChunks(sessionID, chunk); err != nil {
			return err
		}
	}
	return nil
}

// SearchDocuments performs a semantic search on a session's document collection.
func (c *ChromaDBClient) SearchDocuments(sessionID string, query string, topK int) ([]DocumentChunk, error) {
	collectionName := c.getDocumentCollectionName(sessionID)
	var results []DocumentChunk
	err := c.withCollection(collectionName, func(col chroma.Collection) error {
		res, err := col.Query(context.Background(),
			chroma.WithQueryTexts(query),
			chroma.WithNResults(topK),
		)
		if err != nil {
			utility.Logger.WithComponent("chromadb_client").Error(err, "Failed to query documents", "collection_name", collectionName, "query", query, "session_id", sessionID)
			return err
		}

		records := res.ToRecordsGroups()

		for _, group := range records {
			for _, record := range group {
				metadata_doc_id, _ := record.Metadata().GetString("document_id")
				metadata_source_file, _ := record.Metadata().GetString("source_file")
				chunk := DocumentChunk{
					Content:   []byte(record.Document().ContentRaw()),
					ID:        string(record.ID()),
					Embedding: record.Embedding().ContentAsFloat32(),
					Metadata: map[string]any{
						"document_id": metadata_doc_id,
						"source_file": metadata_source_file,
					},
				}
				results = append(results, chunk)
			}
		}
		return nil
	})
	if err != nil {
		utility.Logger.WithComponent("chromadb_client").Error(err, "Failed to search documents", "session_id", sessionID)
		return nil, err
	}

	return results, nil
}

// StoreChatArchive stores archived chat messages in a session's chat collection.
// Takes sessionID and constructs the collection name internally.
func (c *ChromaDBClient) StoreChatArchive(sessionID string, history []template.Message) error {
	collectionName := c.getChatCollectionName(sessionID)
	return c.withCollection(collectionName, func(col chroma.Collection) error {
		// Before embedding each message, we combine the role and the content into one string with a separator,
		// so that the embedding model can capture the role information as well.
		messages := utility.CombineMessageAndRole(history)
		for i, msg := range messages {
			err := col.Add(context.Background(),
				chroma.WithIDs(chroma.DocumentID(fmt.Sprintf("message-%d", i))),
				chroma.WithTexts(msg),
				chroma.WithMetadatas(chroma.NewMetadataFromMap(map[string]any{
					"role":  history[i].Role, // Will still need to use the template here for role extraction.
					"index": i,
				})),
			)
			if err != nil {
				utility.Logger.WithComponent("chromadb_client").Error(err, "Failed to add chat message to archive", "message_index", i, "session_id", sessionID)
				return err
			}
		}
		return nil
	})
}

// SearchChatArchive performs a semantic search on a session's chat archive collection.
// Returns ChatArchiveResult objects since archived chat has different structure than document chunks.
func (c *ChromaDBClient) SearchChatArchive(sessionID string, query string, topK int) ([]ChatArchiveResult, error) {
	collectionName := c.getChatCollectionName(sessionID)
	var results []ChatArchiveResult
	err := c.withCollection(collectionName, func(col chroma.Collection) error {
		res, err := col.Query(context.Background(),
			chroma.WithQueryTexts(query),
			chroma.WithNResults(topK),
		)
		if err != nil {
			utility.Logger.WithComponent("chromadb_client").Error(err, "Failed to query chat archive", "collection_name", collectionName, "query", query, "session_id", sessionID)
			return err
		}

		records := res.ToRecordsGroups()

		for _, group := range records {
			for _, record := range group {
				role, _ := record.Metadata().GetString("role")
				index, _ := record.Metadata().GetInt("index")
				result := ChatArchiveResult{
					Content: record.Document().ContentString(),
					Role:    role,
					Index:   int(index),
				}
				results = append(results, result)
			}
		}
		return nil
	})
	if err != nil {
		utility.Logger.WithComponent("chromadb_client").Error(err, "Failed to search chat archive", "session_id", sessionID)
		return nil, err
	}

	return results, nil
}

// DeleteSessionCollections deletes both document and chat archive collections for a session.
func (c *ChromaDBClient) DeleteSessionCollections(sessionID string) error {
	docCollectionName := c.getDocumentCollectionName(sessionID)
	chatCollectionName := c.getChatCollectionName(sessionID)

	// Delete document collection
	if err := c.deleteCollection(docCollectionName); err != nil {
		utility.Logger.WithComponent("chromadb_client").Error(err, "Failed to delete document collection", "collection_name", docCollectionName, "session_id", sessionID)
		// Continue to delete chat collection even if this fails (eventual consistency)
	}

	// Delete chat archive collection
	if err := c.deleteCollection(chatCollectionName); err != nil {
		utility.Logger.WithComponent("chromadb_client").Error(err, "Failed to delete chat collection", "collection_name", chatCollectionName, "session_id", sessionID)
		return err // Return error from chat collection deletion
	}

	return nil
}

// deleteCollection is a private helper to delete a single collection.
func (c *ChromaDBClient) deleteCollection(collectionName string) error {
	return c.withCollection(collectionName, func(col chroma.Collection) error {
		err := col.Delete(context.Background())
		if err != nil {
			return err
		}
		return nil
	})
}
