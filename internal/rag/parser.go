package rag

import (
	"fmt"
	"local_chatbot/server/utility"
	"os"
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
	if os.IsNotExist(err) {
		utility.Logger.WithComponent("file_checker").Error(err, "File does not exist", "file_path", filePath)
		return false, err
	}

	return checker(fileInfo)
}

func ParsePDF(filePath string) ([]byte, error) {
	verify, err := withFile(filePath, func(fileInfo os.FileInfo) (bool, error) {
		// Implement whether we can parse the file based on its size. For example, we can set a maximum size limit for parsing.
		if fileInfo.Size() > MaxPDFFileSize {
			utility.Logger.WithComponent("pdf_parser").Error(nil, "File size exceeds the maximum limit", "file_path", filePath, "file_size", fileInfo.Size(), "max_size", MaxPDFFileSize)
			return false, fmt.Errorf("File size is over the allowed size limit")
		}
		return true, nil
	})

	if !verify {
		return nil, err
	}

	// Parse the PDF file into byte slice.

	return nil, nil
}

func ParseTxt(filePath string) ([]byte, error) {
	verify, err := withFile(filePath, func(fileInfo os.FileInfo) (bool, error) {
		// Implement whether we can parse the file based on its size. For example, we can set a maximum size limit for parsing.
		if fileInfo.Size() > MaxTXTFileSize {
			utility.Logger.WithComponent("txt_parser").Error(nil, "File size exceeds the maximum limit", "file_path", filePath, "file_size", fileInfo.Size(), "max_size", MaxTXTFileSize)
			return false, fmt.Errorf("File size is over the allowed size limit")
		}
		return true, nil
	})

	if !verify {
		return nil, err
	}

	// Parse the text
	// TODO: Implement the actual parsing logic to read the text file and return its content as a byte slice.

	return nil, nil

}
