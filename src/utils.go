package main

import (
	"encoding/binary"
	"io"
	"os"
)

func readFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func isPEFile(data []byte) bool {
	if data == nil || len(data) < 2 {
		return false
	}

	if data[0] != 'M' || data[1] != 'Z' {
		return false
	}

	if len(data) < 64 {
		return false
	}

	elfanew := binary.LittleEndian.Uint32(data[60:64])
	peHeaderOffset := int(elfanew)

	return binary.LittleEndian.Uint32(data[peHeaderOffset:peHeaderOffset+4]) == 0x4550

}

// getFileName extracts the file name from a given file path.
func getSections(filePath []byte) (string, error) {
	return "not implemented", nil
}

// getFileContents reads and returns the content of the file at the given path.
func getDosHeader(data []byte) (string, error) {
	return "not implemented", nil
}
