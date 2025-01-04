package main

import (
	"debug/pe"
	"encoding/binary"
	"fmt"
	"os"
)

type DOSHeader struct {
	E_magic    uint16     // Magic number
	E_cblp     uint16     // Bytes on last page of file
	E_cp       uint16     // Pages in file
	E_crlc     uint16     // Relocations
	E_cparhdr  uint16     // Size of header in paragraphs
	E_minalloc uint16     // Minimum extra paragraphs needed
	E_maxalloc uint16     // Maximum extra paragraphs needed
	E_ss       uint16     // Initial (relative) SS value
	E_sp       uint16     // Initial SP value
	E_csum     uint16     // Checksum
	E_ip       uint16     // Initial IP value
	E_cs       uint16     // Initial (relative) CS value
	E_lfarlc   uint16     // File address of relocation table
	E_ovno     uint16     // Overlay number
	E_res      [4]uint16  // Reserved words
	E_oemid    uint16     // OEM identifier (for e_oeminfo)
	E_oeminfo  uint16     // OEM information; e_oemid specific
	E_res2     [10]uint16 //Reserved words
	E_ifanew   uint16     // File address of new exe header
	// Additional fields are not always needed but can be added if necessary
}

type PE_FULL struct {
	dos    *DOSHeader // dos header
	peFile *pe.File   // rest of the pe fields
}

var peFull PE_FULL

func parseDOSHeader(filePath string) (*DOSHeader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// The DOS Header is at the beginning of the file
	header := DOSHeader{}
	err = binary.Read(file, binary.LittleEndian, &header)
	if err != nil {
		return nil, fmt.Errorf("failed to read DOS header: %v", err)
	}

	// Verify the magic number
	if header.E_magic != 0x5A4D { // "MZ"
		return nil, fmt.Errorf("invalid DOS header magic number: %#x", header.E_magic)
	}

	return &header, nil
}
