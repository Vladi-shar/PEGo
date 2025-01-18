package main

import (
	"debug/pe"
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
	table, err := createTableForDataDirectories(dataDirs, offset)
	if err != nil {
		displayErrorOnRightPane(ui, err.Error())
		return
	}

	// Replace rightPane with the table
	ui.rightPane.RemoveAll()
	ui.rightPane.Add(table.table)
}

func displaySectionHeadersDetails(ui *MyAppUI, sectionHeaders []*pe.Section, offset uintptr) {
	table, err := createTableForSectionHeaders(sectionHeaders, offset)
	if err != nil {
		displayErrorOnRightPane(ui, err.Error())
		return
	}

	// Replace rightPane with the table
	ui.rightPane.RemoveAll()
	ui.rightPane.Add(table.table)
}
