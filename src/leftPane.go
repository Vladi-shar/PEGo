package main

import (
	"debug/pe"
	"fmt"
)

func getDataDirectories(h any) ([]pe.DataDirectory, error) {
	switch header := h.(type) {
	case *pe.OptionalHeader64:
		return header.DataDirectory[:], nil
	case *pe.OptionalHeader32:
		return header.DataDirectory[:], nil
	default:
		return nil, fmt.Errorf("unknown header type")
	}
}

func getOptionalHeader(peFile *pe.File) (any, error) {
	switch peFile.FileHeader.Machine {
	case pe.IMAGE_FILE_MACHINE_AMD64:
		// 64-bit binary
		return peFile.OptionalHeader.(*pe.OptionalHeader64), nil
	case pe.IMAGE_FILE_MACHINE_I386:
		// 32-bit binary
		return peFile.OptionalHeader.(*pe.OptionalHeader32), nil
	default:
		// Unsupported or unknown architecture
		return nil, fmt.Errorf("unsupported Machine type: 0x%x", peFile.FileHeader.Machine)
	}
}

func getPeTreeMap(peFile *pe.File, filePath string) map[string][]string {
	data := map[string][]string{}
	root := "File: " + filePath

	data[""] = []string{root}
	data[root] = []string{"Dos Header", "Nt Headers", "Section Headers"}
	data["Nt Headers"] = []string{"File Header", "Optional Header"}
	data["Optional Header"] = []string{"Data Directories"}

	// Access the Optional Header
	optHeader, err := getOptionalHeader(peFile)
	if err != nil {
		fmt.Println(err)
		return map[string][]string{}
	}

	// Loop through the Data Directories
	fmt.Println("Data Directories:")
	dataDirs, err := getDataDirectories(optHeader)
	if err != nil {
		fmt.Println(err)
		return map[string][]string{}
	}

	for i, dir := range dataDirs {
		if i < len(directoryNames) {
			fmt.Printf("Directory [%d] %s: VirtualAddress=0x%x, Size=%x\n", i, directoryNames[i], dir.VirtualAddress, dir.Size)

			if dir.VirtualAddress != 0 && dir.Size != 0 {
				// data[directoryNames[i]] = []string{}
				data[root] = append(data[root], directoryNames[i])
			}
		} else {
			fmt.Printf("Directory [%d]: UNKNOWN VirtualAddress=0x%x, Size=%d\n", i, dir.VirtualAddress, dir.Size)
		}
	}

	return data
}
