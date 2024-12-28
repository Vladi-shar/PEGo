package main

import (
	"io"
	"os"
	"path/filepath"
)

// getFileName extracts the file name from a given file path.
func getFileName(filePath string) string {
	return filepath.Base(filePath)
}

// getFileContents reads and returns the content of the file at the given path.
func getFileContents(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
