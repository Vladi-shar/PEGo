package main

import (
	"debug/pe"
	"fmt"
)

func getPeTreeMap(peFile *pe.File, filePath string) map[string][]string {
	data := map[string][]string{}
	root := "File: " + filePath

	data[""] = []string{root}
	data[root] = []string{"Dos Header", "Nt Headers", "Section Headers"}
	data["Nt Headers"] = []string{"File Header", "Optional Header"}
	data["Optional Header"] = []string{"Data Directories"}

	// Access the Optional Header
	optHeader, ok := peFile.OptionalHeader.(*pe.OptionalHeader64) // Use OptionalHeader64 for 64-bit PE files
	if !ok {
		fmt.Println("Failed to cast optional header")
		return map[string][]string{}
	}

	directoryNames := []string{
		"Export Directory",
		"Import Directory",
		"Resource Directory",
		"Exception Directory",
		"Security Directory",
		"Base Relocation Table",
		"Debug Directory",
		"Architecture Specific Data",
		"RVA of GlobalPtr",
		"TLS Directory",
		"Load Configuration Directory",
		"Bound Import Directory",
		"Import Address Table",
		"Delay Load Import Descriptors",
		".NET Header",
	}

	// Loop through the Data Directories
	fmt.Println("Data Directories:")
	for i, dir := range optHeader.DataDirectory {
		if i < len(directoryNames) {
			fmt.Printf("Directory [%d] %s: VirtualAddress=0x%x, Size=%d\n", i, directoryNames[i], dir.VirtualAddress, dir.Size)

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
