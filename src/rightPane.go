package main

import (
	"bytes"
	"debug/pe"
	"encoding/binary"
	"fmt"

	"fyne.io/fyne/v2/container"
)

func displayFileProperties(ui *MyAppUI, fileProperties FileProperties) {
	// Remove all widgets from the right pane
	ui.rightPane.RemoveAll()
	propertiesTable, err := createTableForProperties(fileProperties)
	if err != nil {
		displayErrorOnRightPane(ui, err.Error())
		return
	}
	resourcesTable, err := createTableForResources(fileProperties.FileResources)
	if err != nil {
		displayErrorOnRightPane(ui, err.Error())
		return
	}
	ui.rightPane.RemoveAll()

	split := container.NewVSplit(propertiesTable.table, resourcesTable.table)
	ui.rightPane.Add(split)
}

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

func displayExportTableDetails(ui *MyAppUI, exportDir pe.DataDirectory, peFile *pe.File, fileData []byte) {

	// exportDir.VirtualAddress
	var exportHeader IMAGE_EXPORT_DIRECTORY

	// read from osFile at ExportDir.VirtualAddress into exportHeader
	// if _, err := osFile.Seek(int64(exportDir.VirtualAddress), io.SeekStart); err != nil {
	// 	displayErrorOnRightPane(ui, err.Error())
	// 	return
	// }
	fmt.Printf("virtual address: 0x%x\n", exportDir.VirtualAddress)
	fmt.Printf("file size: %d\n", len(fileData))
	exportDirRawOffset, err := rvaToOffset(peFile, exportDir.VirtualAddress)
	if err != nil {
		displayErrorOnRightPane(ui, err.Error())
		return
	}

	reader := bytes.NewReader(fileData[exportDirRawOffset:])
	if err := binary.Read(reader, binary.LittleEndian, &exportHeader); err != nil {
		displayErrorOnRightPane(ui, err.Error())
		return
	}
	table, err := createTableFromStruct(exportHeader, uintptr(exportDirRawOffset))
	if err != nil {
		displayErrorOnRightPane(ui, err.Error())
		return
	}

	table2, err := createTableForExports(peFile, fileData, exportHeader)
	if err != nil {
		displayErrorOnRightPane(ui, err.Error())
		return
	}

	split := container.NewVSplit(table.table, table2.table)

	ui.rightPane.RemoveAll()
	ui.rightPane.Add(split)
}
