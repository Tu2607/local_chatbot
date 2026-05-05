package rag

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"local_chatbot/server/utility"
	"os"
	"os/exec"
)

type SizeLimitChecker func(fileInfo os.FileInfo) (bool, error)

const (
	MaxTXTFileSize    = 5 * 1024 * 1024  // 5MB
	MaxPDFFileSize    = 10 * 1024 * 1024 // 10MB
	MaxDOCXFileSize   = 10 * 1024 * 1024 // 10MB
	ChunkSizeBytes    = 8192             // 8KB chunks
	ChunkOverlapBytes = 512              // 512 byte overlap
)

func withFile(filePath string, checker SizeLimitChecker) (bool, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			utility.Logger.WithComponent("file_checker").Error(err, "File does not exist", "file_path", filePath)
			return false, err
		}

		utility.Logger.WithComponent("file_checker").Error(err, "Failed to stat file", "file_path", filePath)
		return false, err
	}

	return checker(fileInfo)
}

func ParsePDF(filePath string) ([]byte, error) {
	verify, err := withFile(filePath, func(fileInfo os.FileInfo) (bool, error) {
		if fileInfo.Size() > MaxPDFFileSize {
			utility.Logger.WithComponent("pdf_parser").Warn("File size exceeds the maximum limit", "file_path", filePath, "file_size", fileInfo.Size(), "max_size", MaxPDFFileSize)
			return false, fmt.Errorf("File size is over the allowed size limit")
		}
		return true, nil
	})

	if !verify {
		return nil, err
	}

	// Verify if the pdftotext cli is available in the system. If not, return an error.
	_, err = exec.LookPath("pdftotext")
	if err != nil {
		utility.Logger.WithComponent("pdf_parser").Error(err, "pdftotext command not found. Please ensure poppler-utils is installed and pdftotext is in the system PATH.")
		return nil, fmt.Errorf("pdftotext command not found. Please ensure poppler-utils is installed and pdftotext is in the system PATH")
	}

	// Use the pdftotext command to extract text from the PDF file.
	cmd := exec.Command("pdftotext", filePath, "-")
	output, err := cmd.Output()
	if err != nil {
		utility.Logger.WithComponent("pdf_parser").Error(err, "Failed to extract text from PDF file using pdftotext", "file_path", filePath)
		return nil, err
	}

	return output, nil
}

func ParseTxt(filePath string) ([]byte, error) {
	verify, err := withFile(filePath, func(fileInfo os.FileInfo) (bool, error) {
		if fileInfo.Size() > MaxTXTFileSize {
			utility.Logger.WithComponent("txt_parser").Warn("File size exceeds the maximum limit", "file_path", filePath, "file_size", fileInfo.Size(), "max_size", MaxTXTFileSize)
			return false, fmt.Errorf("File size is over the allowed size limit")
		}
		return true, nil
	})

	if !verify {
		return nil, err
	}

	// Parse the text
	file, _ := os.Open(filePath)
	defer file.Close()

	// Apply a strict buffer size to read the file
	buf := make([]byte, MaxTXTFileSize)

	n_bytes, err := file.Read(buf)

	// Empty files will return EOF, which is not an error in this context, so we should only log actual errors.
	if err != nil && err != io.EOF {
		utility.Logger.WithComponent("txt_parser").Error(err, "Failed to read text file", "file_path", filePath)
		return nil, err
	}

	return buf[:n_bytes], nil

}

func ParseDocx(filePath string) ([]byte, error) {
	// For simplicity, we can treat docx files as zip files and extract the text content from the relevant XML files inside. However, this is a simplified approach and may not cover all edge cases or formatting. For a more robust solution, consider using a dedicated library for parsing docx files.
	verify, err := withFile(filePath, func(fileInfo os.FileInfo) (bool, error) {
		if fileInfo.Size() > MaxDOCXFileSize {
			utility.Logger.WithComponent("docx_parser").Warn("File size exceeds the maximum limit", "file_path", filePath, "file_size", fileInfo.Size(), "max_size", MaxDOCXFileSize)
			return false, fmt.Errorf("File size is over the allowed size limit")
		}
		return true, nil
	})

	if !verify {
		return nil, err
	}

	// Open the docx file as a zip archive
	r, err := zip.OpenReader(filePath)
	if err != nil {
		utility.Logger.WithComponent("docx_parser").Error(err, "Failed to open docx file as zip archive")
		return nil, err
	}
	defer r.Close()

	// Iterate through the files in the zip archive to find the document.xml file which contains the main text content
	var xmlFile io.ReadCloser
	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			xmlFile, err = f.Open()
			if err != nil {
				utility.Logger.WithComponent("docx_parser").Error(err, "Failed to open document.xml from docx archive")
				return nil, err
			}
			break
		}
	}

	if xmlFile == nil {
		utility.Logger.WithComponent("docx_parser").Error(fmt.Errorf("document.xml not found in docx archive"), "Failed to find document.xml in docx file")
		return nil, fmt.Errorf("document.xml not found in docx file")
	}

	defer xmlFile.Close()

	// Define a struct to unmarshal the XML content.
	type DocumentTextXML struct {
		XMLName xml.Name `xml:"w:document"`
		Body    struct {
			Paragraphs []struct {
				Runs []struct {
					Text string `xml:"t"`
				} `xml:"r"`
			} `xml:"p"`
		} `xml:"body"`
	}

	// Decode the XML content to extract the text
	var docText DocumentTextXML
	decoder := xml.NewDecoder(xmlFile)
	err = decoder.Decode(&docText)
	if err != nil {
		utility.Logger.WithComponent("docx_parser").Error(err, "Failed to decode XML content from document.xml")
		return nil, err
	}

	content := make([]byte, MaxDOCXFileSize)
	textBytesSlices := make([][]byte, 0)

	// TODO: Get all the text content
	for _, paragraph := range docText.Body.Paragraphs {
		for _, run := range paragraph.Runs {
			textBytesSlices = append(textBytesSlices, []byte(run.Text))
		}
		textBytesSlices = append(textBytesSlices, []byte("\n")) // Add a newline after each paragraph
	}

	// Convert [][]byte to io.reader.
	readers := make([]io.Reader, len(textBytesSlices))
	for i, reader := range textBytesSlices {
		readers[i] = bytes.NewReader(reader)
	}

	bytesRead, err := io.MultiReader(readers...).Read(content)
	if err != nil && err != io.EOF {
		utility.Logger.WithComponent("docx_parser").Error(err, "Failed to read combined text content from document.xml")
		return nil, err
	}
	utility.Logger.WithComponent("docx_parser").Debug("Successfully parsed docx file", "file_path", filePath, "bytes_read", bytesRead)

	return content[:bytesRead], nil
}

// ChunkParsedContent splits content into overlapping chunks of ChunkSizeBytes.
// It respects UTF-8 boundaries to avoid splitting multi-byte characters.
// If overlapSize is invalid, it defaults to ChunkOverlapBytes.
func ChunkParsedContent(content []byte, overlapSize int) [][]byte {
	if len(content) == 0 {
		return [][]byte{}
	}

	// Validate and normalize overlapSize
	if overlapSize < 0 || overlapSize >= ChunkSizeBytes {
		utility.Logger.WithComponent("chunking").Warn("Invalid overlap size", "requested", overlapSize, "using_default", ChunkOverlapBytes)
		overlapSize = ChunkOverlapBytes
	}

	size := len(content)

	// If content is smaller than chunk size, return as a single chunk
	if size <= ChunkSizeBytes {
		return [][]byte{content}
	}

	// Keep it simple and just define an array of byte slices to hold the chunks.
	// No need for predicting the number of chunks in advance, we can just append to the slice as we go.
	var chunks [][]byte

	for start := 0; start < size; {
		end := min(start+ChunkSizeBytes, size)
		end = utility.PreviousUTF8Boundary(content, end)

		if end <= start {
			break
		}

		chunks = append(chunks, content[start:end])

		// Move to next chunk position with overlap
		if end == size {
			break
		}
		start = utility.NextUTF8Boundary(content, end-overlapSize)
	}

	utility.Logger.WithComponent("rag_parser").Debug("Chunking completed", "total_chunks", len(chunks), "chunk_size_bytes", ChunkSizeBytes, "overlap_size_bytes", overlapSize)
	return chunks
}
