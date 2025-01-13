package main

import "debug/pe"

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

	// Replace rightPane with the table
	ui.rightPane.RemoveAll()
	ui.rightPane.Add(table.table)

}
