package main

import (
	"bytes"
	"debug/pe"
	"encoding/binary"
	"fmt"
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
	E_res2     [10]uint16 // Reserved words
	E_ifanew   uint32     // File address of new exe header
	// Additional fields are not always needed but can be added if necessary
}

type NtHeaders struct {
	Signature uint32
}

type PeFull struct {
	dos      *DOSHeader // dos header
	nt       *NtHeaders // nt headers
	peFile   *pe.File   // rest of the pe fields
	fileData []byte     // raw file
}

type IMAGE_EXPORT_DIRECTORY struct {
	Characteristics       uint32
	TimeDateStamp         uint32
	MajorVersion          uint16
	MinorVersion          uint16
	Name                  uint32
	Base                  uint32
	NumberOfFunctions     uint32
	NumberOfNames         uint32
	AddressOfFunctions    uint32
	AddressOfNames        uint32
	AddressOfNameOrdinals uint32
}

var directoryNames = []string{
	"Export Table",
	"Import Table",
	"Resource Table",
	"Exception Table",
	"Certificate Table",
	"Base Relocation Table",
	"Debug",
	"Architecture",
	"Global Ptr",
	"TLS Table",
	"Load Config Table",
	"Bound Import",
	"IAT",
	"Delay Import Descriptor",
	"CLR Runtime Header",
	"Reserved",
}

func NewPeFull(_dos *DOSHeader, _nt *NtHeaders, _peFile *pe.File, _fileData []byte) *PeFull {
	return &PeFull{
		dos:      _dos,
		nt:       _nt,
		peFile:   _peFile,
		fileData: _fileData,
	}
}

func parseDOSHeader(fileData []byte) (*DOSHeader, error) {

	// The DOS Header is at the beginning of the file
	header := DOSHeader{}
	reader := bytes.NewReader(fileData[0:])
	err := binary.Read(reader, binary.LittleEndian, &header)
	if err != nil {
		return nil, fmt.Errorf("failed to read DOS header: %v", err)
	}
	// Verify the magic number
	if header.E_magic != 0x5A4D { // "MZ"
		return nil, fmt.Errorf("invalid DOS header magic number: %#x", header.E_magic)
	}

	return &header, nil
}

func parseNtHeaders(fileData []byte, dos *DOSHeader) (*NtHeaders, error) {

	// The NT Headers are a signature followed by the rest of the headers
	headers := NtHeaders{}
	reader := bytes.NewReader(fileData[int64(dos.E_ifanew):])
	err := binary.Read(reader, binary.LittleEndian, &headers)
	if err != nil {
		return nil, fmt.Errorf("failed to read NT Headers: %v", err)
	}

	// Verify the signature
	if headers.Signature != 0x00004550 { // "PE\0\0"
		return nil, fmt.Errorf("invalid NT Headers signature: %#x", headers.Signature)
	}

	return &headers, nil
}

func rvaToOffset(pe *pe.File, rva uint32) (uint32, error) {
	for _, sh := range pe.Sections {
		size := sh.VirtualSize
		// Some tools pad VirtualSize to multiple of FileAlignment; make sure to handle that.
		// But for simplicity, let's just use VirtualSize as is:
		if rva >= sh.VirtualAddress && rva < sh.VirtualAddress+size {
			delta := rva - sh.VirtualAddress
			fileOffset := sh.Offset + delta
			// if fileOffset > uint32(pe.FileSize) {
			// 	return 0, fmt.Errorf("file offset 0x%X is out of file bounds", fileOffset)
			// }
			return fileOffset, nil
		}
	}
	return 0, fmt.Errorf("RVA 0x%X not found in any section", rva)
}
