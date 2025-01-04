package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

// DOS header
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

func displayDosHeaderDetails(ui *MyAppUI, filePath string) {
	// Format the DOS header fields
	dosHeader, err := parseDOSHeader(filePath)

	if err != nil {
		ui.rightPane.SetText("Failed to parse DOS header")
		return
	}

	details := fmt.Sprintf(
		`DOS Header:
	e_magic: %#x
	e_cblp: %#x
	e_cp: %#x
	e_crlc: %#x
	e_cparhdr: %#x
	e_minalloc: %#x
	e_maxalloc: %#x
	e_ss: %#x
	e_sp: %#x
	e_csum: %#x
	e_ip: %#x
	e_cs: %#x
	e_lfarlc: %#x
	e_ovno: %#x
	e_res: %#x
	e_oemid: %#x
	e_oeminfo: %#x
	e_res2: %#x    
	e_ifanew: %#x`,
		dosHeader.E_magic,
		dosHeader.E_cblp,
		dosHeader.E_cp,
		dosHeader.E_crlc,
		dosHeader.E_cparhdr,
		dosHeader.E_minalloc,
		dosHeader.E_maxalloc,
		dosHeader.E_ss,
		dosHeader.E_sp,
		dosHeader.E_csum,
		dosHeader.E_ip,
		dosHeader.E_cs,
		dosHeader.E_lfarlc,
		dosHeader.E_ovno,
		dosHeader.E_res,
		dosHeader.E_oemid,
		dosHeader.E_oeminfo,
		dosHeader.E_res2,
		dosHeader.E_ifanew,
	)

	// Display the details in the right pane
	ui.rightPane.SetText(details)
}
