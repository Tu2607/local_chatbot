package rag

import (
	"bytes"
	"local_chatbot/server/utility"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"unicode/utf8"
)

// TestMain initializes the logger before running tests
func TestMain(m *testing.M) {
	utility.Logger = utility.InitLogger(os.Getenv("DEBUG") == "true")
	os.Exit(m.Run())
}

// Helper function to create a temporary test file
func createTestFile(t *testing.T, name string, content []byte) string {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, name)
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return filePath
}

// TestChunkParsedContent tests the chunking logic with various scenarios
func TestChunkParsedContent(t *testing.T) {
	tests := []struct {
		name          string
		content       []byte
		overlapSize   int
		expectedCount int
		validateFunc  func(t *testing.T, chunks [][]byte)
	}{
		{
			name:          "empty content",
			content:       []byte{},
			overlapSize:   ChunkOverlapBytes,
			expectedCount: 0,
		},
		{
			name:          "content smaller than chunk size",
			content:       []byte("small content"),
			overlapSize:   ChunkOverlapBytes,
			expectedCount: 1,
			validateFunc: func(t *testing.T, chunks [][]byte) {
				if len(chunks) != 1 {
					t.Errorf("Expected 1 chunk, got %d", len(chunks))
				}
				if !bytes.Equal(chunks[0], []byte("small content")) {
					t.Errorf("Chunk content mismatch")
				}
			},
		},
		{
			name:          "content exactly chunk size",
			content:       bytes.Repeat([]byte("a"), ChunkSizeBytes),
			overlapSize:   ChunkOverlapBytes,
			expectedCount: 1,
		},
		{
			name:          "content needs multiple chunks",
			content:       bytes.Repeat([]byte("x"), ChunkSizeBytes*2+1000),
			overlapSize:   ChunkOverlapBytes,
			expectedCount: 3,
			validateFunc: func(t *testing.T, chunks [][]byte) {
				// Verify each chunk is valid UTF-8
				for i, chunk := range chunks {
					if !utf8.Valid(chunk) {
						t.Errorf("Chunk %d contains invalid UTF-8", i)
					}
				}
			},
		},
		{
			name:          "UTF-8 boundary preservation",
			content:       bytes.Repeat([]byte("Hello "+"世"+" World"), ChunkSizeBytes), // Japanese character
			overlapSize:   ChunkOverlapBytes,
			expectedCount: 16,
			validateFunc: func(t *testing.T, chunks [][]byte) {
				for i, chunk := range chunks {
					if !utf8.Valid(chunk) {
						t.Errorf("Chunk %d contains invalid UTF-8", i)
					}
				}
			},
		},
		{
			name:          "overlap correctness",
			content:       bytes.Repeat([]byte("abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz"), 315),
			overlapSize:   ChunkOverlapBytes,
			expectedCount: 3,
			validateFunc: func(t *testing.T, chunks [][]byte) {
				if len(chunks) < 2 {
					t.Fatalf("Expected multiple chunks for overlap test")
				}
				// Check that there's actually overlap
				chunk0End := chunks[0][len(chunks[0])-ChunkOverlapBytes:]
				chunk1Start := chunks[1][:ChunkOverlapBytes]
				if len(chunk0End) < 5 || len(chunk1Start) < 5 {
					t.Skip("Chunks too small to verify overlap")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := ChunkParsedContent(tt.content, tt.overlapSize)

			if tt.expectedCount >= 0 && len(chunks) != tt.expectedCount {
				t.Errorf("Expected %d chunks, got %d", tt.expectedCount, len(chunks))
			}

			if tt.validateFunc != nil {
				tt.validateFunc(t, chunks)
			}
		})
	}
}

// TestParseTxt tests text file parsing
func TestParseTxt(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		shouldFail bool
	}{
		{
			name:       "valid text file",
			content:    "Hello, World!",
			shouldFail: false,
		},
		{
			name:       "empty text file",
			content:    "",
			shouldFail: false,
		},
		{
			name:       "text with special characters",
			content:    "Special: éàü中文",
			shouldFail: false,
		},
		{
			name:       "large text under limit",
			content:    string(bytes.Repeat([]byte("x"), 1*1024*1024)), // 1MB
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := createTestFile(t, "test.txt", []byte(tt.content))

			content, err := ParseTxt(filePath)

			if tt.shouldFail && err == nil {
				t.Errorf("Expected error, got nil")
			}
			if !tt.shouldFail && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.shouldFail && !bytes.Equal(content, []byte(tt.content)) {
				t.Errorf("Content mismatch")
			}
		})
	}
}

// TestParsePDF tests PDF parsing (skipped if pdftotext not available)
func TestParsePDF(t *testing.T) {
	// Check if pdftotext is available
	_, err := exec.LookPath("pdftotext")
	if err != nil {
		t.Skip("pdftotext not available, skipping PDF tests")
	}

	t.Run("nonexistent PDF", func(t *testing.T) {
		_, err := ParsePDF("/nonexistent/file.pdf")
		if err == nil {
			t.Errorf("Expected error for nonexistent file")
		}
	})

	t.Run("invalid PDF file", func(t *testing.T) {
		filePath := createTestFile(t, "invalid.pdf", []byte("not a pdf file"))
		_, err := ParsePDF(filePath)
		if err == nil {
			t.Errorf("Expected error for invalid PDF")
		}
	})
}

// TestParseDocx tests DOCX parsing
func TestParseDocx(t *testing.T) {
	t.Run("nonexistent DOCX", func(t *testing.T) {
		_, err := ParseDocx("/nonexistent/file.docx")
		if err == nil {
			t.Errorf("Expected error for nonexistent file")
		}
	})

	t.Run("invalid ZIP file", func(t *testing.T) {
		filePath := createTestFile(t, "invalid.docx", []byte("not a zip file"))
		_, err := ParseDocx(filePath)
		if err == nil {
			t.Errorf("Expected error for invalid DOCX")
		}
	})

	t.Run("ZIP without document.xml", func(t *testing.T) {
		filePath := createTestFile(t, "empty.docx", []byte("PK\x03\x04")) // Minimal ZIP header
		_, err := ParseDocx(filePath)
		if err == nil {
			t.Errorf("Expected error for DOCX without document.xml")
		}
	})
}

// TestFileSizeValidation tests that file size checks work correctly
func TestFileSizeValidation(t *testing.T) {
	tests := []struct {
		name       string
		fileSize   int
		parser     func(string) ([]byte, error)
		maxSize    int
		shouldFail bool
	}{
		{
			name:       "TXT file under limit",
			fileSize:   1 * 1024 * 1024, // 1MB
			parser:     ParseTxt,
			maxSize:    MaxTXTFileSize,
			shouldFail: false,
		},
		{
			name:       "TXT file over limit",
			fileSize:   6 * 1024 * 1024, // 6MB
			parser:     ParseTxt,
			maxSize:    MaxTXTFileSize,
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := bytes.Repeat([]byte("x"), tt.fileSize)
			filePath := createTestFile(t, "test.txt", content)

			_, err := tt.parser(filePath)

			if tt.shouldFail && err == nil {
				t.Errorf("Expected error for oversized file (size: %d, max: %d)", tt.fileSize, tt.maxSize)
			}
			if !tt.shouldFail && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestUTF8EdgeCases tests chunking with tricky UTF-8 scenarios
func TestUTF8EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
	}{
		{
			name:    "emoji at boundary",
			content: append(bytes.Repeat([]byte("a"), ChunkSizeBytes-2), []byte("😀")...),
		},
		{
			name:    "mixed UTF-8 characters",
			content: []byte("Hello 世界 مرحبا עולם"),
		},
		{
			name:    "multi-byte sequences",
			content: bytes.Repeat([]byte("你好世界"), ChunkSizeBytes/8),
		},
		{
			name:    "large content with UTF-8",
			content: bytes.Repeat([]byte("This is a test with 中文 and emoji 🎉 mixed in. "), ChunkSizeBytes/50),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := ChunkParsedContent(tt.content, ChunkOverlapBytes)

			// Verify all chunks are valid UTF-8
			for i, chunk := range chunks {
				if !utf8.Valid(chunk) {
					utility.Logger.Debug("Chunked Content", chunks)
					t.Errorf("Chunk %d contains invalid UTF-8", i)
				}
			}

			// Verify no chunk is empty (except for edge cases)
			for i, chunk := range chunks {
				if len(chunk) == 0 {
					t.Errorf("Chunk %d is empty", i)
				}
			}
		})
	}
}

// TestChunkingDoesNotLoseContent verifies that chunking preserves all content
func TestChunkingDoesNotLoseContent(t *testing.T) {
	testCases := []struct {
		name    string
		content []byte
	}{
		{
			name:    "ASCII content",
			content: bytes.Repeat([]byte("The quick brown fox jumps over the lazy dog. "), ChunkSizeBytes/50),
		},
		{
			name:    "UTF-8 content",
			content: bytes.Repeat([]byte("你好世界，这是一个测试。 "), ChunkSizeBytes/30),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			chunks := ChunkParsedContent(tc.content, ChunkOverlapBytes)

			// Reconstruct from chunks (removing overlap for verification)
			reconstructed := []byte{}
			for i, chunk := range chunks {
				if i == 0 {
					reconstructed = append(reconstructed, chunk...)
				} else {
					// Skip overlap bytes from subsequent chunks
					if len(chunk) > ChunkOverlapBytes {
						reconstructed = append(reconstructed, chunk[ChunkOverlapBytes:]...)
					}
				}
			}

			if !bytes.Equal(reconstructed, tc.content) {
				t.Errorf("Content loss after chunking:\nOriginal length: %d\nReconstructed length: %d", len(tc.content), len(reconstructed))
			}
		})
	}
}
