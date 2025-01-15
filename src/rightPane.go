package main

import (
	"debug/pe"
	"fmt"
)

func displayDosHeaderDetails(ui *MyAppUI, dosHeader *DOSHeader, offset uintptr) {

	table, err := createTableFromStruct(dosHeader, offset)
	if err != nil {
		displayErrorOnRightPane(ui, err.Error())
		return
	}

	// Replace rightPane with the table
	ui.rightPane.RemoveAll()
	ui.rightPane.Add(table.table)

}

func displayNtHeadersDetails(ui *MyAppUI, ntHeaders *NtHeaders, offset uintptr) {

	table, err := createTableFromStruct(ntHeaders, offset)
	if err != nil {
		displayErrorOnRightPane(ui, err.Error())
		return
	}

	// Replace rightPane with the table
	ui.rightPane.RemoveAll()
	ui.rightPane.Add(table.table)

}

func displayFileHeaderDetails(ui *MyAppUI, fileHeader *pe.FileHeader, offset uintptr) {

	table, err := createTableFromStruct(fileHeader, offset)
	if err != nil {
		displayErrorOnRightPane(ui, err.Error())
		return
	}

	// Replace rightPane with the table
	ui.rightPane.RemoveAll()
	ui.rightPane.Add(table.table)

}

func displayOptionalHeaderDetails(ui *MyAppUI, optHeader any, offset uintptr) {

	table, err := createTableFromStruct(optHeader, offset)
	if err != nil {
		displayErrorOnRightPane(ui, err.Error())
		return
	}

	// Remove DataDirectories row
	table.removeRow(len(table.data) - 1)

	// Replace rightPane with the table
	ui.rightPane.RemoveAll()
	ui.rightPane.Add(table.table)

}

func displayDataDirectoryDetails(ui *MyAppUI, dataDirs []pe.DataDirectory, offset uintptr) {

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

	// Replace rightPane with the table
	ui.rightPane.RemoveAll()
	ui.rightPane.Add(table.table)

}
