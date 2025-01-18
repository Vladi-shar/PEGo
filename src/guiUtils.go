package main

import (
	"bytes"
	"debug/pe"
	_ "embed"
	"encoding/binary"
	"fmt"
	"image/png"
	"os"
	"reflect"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/sqweek/dialog"
)

//go:embed winres\\logosmall.png
var iconBytes []byte

var MyApp fyne.App

type MyAppUI struct {
	leftPane  *fyne.Container
	rightPane *fyne.Container
}

func initUIElements() *MyAppUI {
	return &MyAppUI{
		leftPane:  container.NewStack(),
		rightPane: container.NewStack(),
	}
}

func loadIcon() fyne.Resource {
	_, err := png.Decode(bytes.NewReader(iconBytes))
	if err != nil {
		panic("Failed to load embedded icon: " + err.Error())
	}
	return fyne.NewStaticResource("logosmall.png", iconBytes)
}

func displayPopup(heading string, msg string) {

	if MyApp == nil {
		fmt.Println("MyApp is not initialized")
		return
	}

	popupWindow := MyApp.NewWindow(heading)
	popupLabel := widget.NewLabel(msg)

	popupWindow.SetContent(container.NewVBox(
		popupLabel,
		widget.NewButton("OK", func() {
			popupWindow.Close()
		}),
	))
	popupWindow.Resize(fyne.NewSize(250, 150))
	popupWindow.Show()
}

func displayErrorOnRightPane(ui *MyAppUI, msg string) {
	ui.rightPane.RemoveAll()
	ui.rightPane.Add(widget.NewLabel(msg))
}

func InitPaneView(window fyne.Window) {
	// Create two panes
	ui := initUIElements()
	// Define the initial data for the tree
	data := map[string][]string{}
	// Create the tree widget
	tree := widget.NewTree(
		// Define the child nodes for each node
		func(uid widget.TreeNodeID) []widget.TreeNodeID {
			return data[uid]
		},
		// Define whether a node is a branch
		func(uid widget.TreeNodeID) bool {
			_, isBranch := data[uid]
			return isBranch
		},
		// Define how to create the template for branches and leaves
		func(branch bool) fyne.CanvasObject {
			txt := canvas.NewText("", nil)
			txt.TextSize = 12
			return txt
		},
		// Define how to update the template for a specific node
		func(uid widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
			txt := obj.(*canvas.Text)
			txt.Text = uid
			txt.Refresh()
		},
	)

	ui.leftPane.Add(tree)
	var filePath string

	var peFull *PeFull

	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Open", func() {
			var err error
			filePath, err = dialog.File().Title("Select a File").Load()
			if err != nil {
				if err.Error() != "cancelled" { // Ignore "cancelled" error
					fmt.Println("Error opening file:", err)
				}
				return
			}

			file, err := os.Open(filePath)
			if err != nil {
				errorMessage := fmt.Sprintf("Error opening file: %v", err)
				fmt.Println(errorMessage)
			}
			peFile, err := pe.NewFile(file)
			if err != nil {
				file.Close()
				errorMessage := fmt.Sprintf("Error opening file: %v", err)
				displayErrorOnRightPane(ui, "Unsupported file format")
				fmt.Println(errorMessage)
				return
			}
			defer file.Close()

			dos, err := parseDOSHeader(file)
			if err != nil {
				errorMessage := fmt.Sprintf("Error parsing dos header: %v", err)
				displayErrorOnRightPane(ui, "Error parsing dos header")
				fmt.Println(errorMessage)
				return
			}

			nt, err := parseNtHeaders(file, dos)
			if err != nil {
				errorMessage := fmt.Sprintf("Error parsing nt headers: %v", err)
				displayErrorOnRightPane(ui, "Error parsing nt headers")
				fmt.Println(errorMessage)
				return
			}

			peFull = NewPeFull(dos, nt, peFile)
			data = getPeTreeMap(peFile, filePath)

			tree.Root = "File: " + filePath
			tree.Refresh()
			tree.OpenAllBranches()
		}),
	)

	tree.OnSelected = func(uid widget.TreeNodeID) {
		switch uid {
		case "Dos Header":
			// Call the function to display DOS header details
			displayDosHeaderDetails(ui, peFull.dos, 0)
		case "Nt Headers":
			displayNtHeadersDetails(ui, peFull.nt, uintptr(peFull.dos.E_ifanew))
		case "File Header":
			// fmt.Printf("sizeof nt: %d\n", unsafe.Sizeof(peFull.nt))
			// fmt.Printf("sizeof nt.signature: %d\n", unsafe.Sizeof(peFull.nt.Signature))
			displayFileHeaderDetails(ui, &peFull.peFile.FileHeader, uintptr(peFull.dos.E_ifanew)+uintptr(binary.Size(peFull.nt)))
		case "Optional Header":
			optHeader, err := getOptionalHeader(peFull.peFile)
			if err != nil {
				displayErrorOnRightPane(ui, err.Error())
				return
			}
			displayOptionalHeaderDetails(ui, optHeader, uintptr(peFull.dos.E_ifanew)+uintptr(binary.Size(peFull.nt))+uintptr(binary.Size(peFull.peFile.FileHeader)))
		case "Data Directories":
			optHeader, err := getOptionalHeader(peFull.peFile)
			if err != nil {
				displayErrorOnRightPane(ui, err.Error())
				return
			}
			dataDirs, err := getDataDirectories(optHeader)
			if err != nil {
				displayErrorOnRightPane(ui, err.Error())
				return
			}

			displayDataDirectoryDetails(ui, dataDirs, uintptr(peFull.dos.E_ifanew)+uintptr(binary.Size(peFull.nt))+uintptr(binary.Size(peFull.peFile.FileHeader))+uintptr(binary.Size(peFull.peFile.OptionalHeader))-uintptr(binary.Size(dataDirs)))

		case "Section Headers":
			displaySectionHeadersDetails(ui, peFull.peFile.Sections, uintptr(peFull.dos.E_ifanew)+uintptr(binary.Size(peFull.nt))+uintptr(binary.Size(peFull.peFile.FileHeader))+uintptr(binary.Size(peFull.peFile.OptionalHeader)))
		default:
			ui.rightPane.RemoveAll()
			ui.rightPane.Add(widget.NewLabel(uid))

		}

	}

	// Create the main menu
	mainMenu := fyne.NewMainMenu(fileMenu)

	// Create a horizontal split
	split := container.NewHSplit(ui.leftPane, ui.rightPane)
	split.SetOffset(0.3) // Set the split ratio (0.5 means equal halves)
	fixedSplit := container.NewStack(split)

	// Set the content with a vertical layout
	content := container.NewBorder(nil, nil, nil, nil, fixedSplit)

	// Set the menu and content in the window
	window.SetMainMenu(mainMenu)
	window.SetContent(content)

	// Show and run the application
	window.Resize(fyne.NewSize(800, 600))
	window.ShowAndRun()
}

func createTableFromStruct(header any, offset uintptr) (*sortableTable, error) {
	// Use reflection to iterate over the struct fields
	t := reflect.TypeOf(header)
	v := reflect.ValueOf(header)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct or a pointer to a struct, got %s", t.Kind())
	}

	// Prepare the data slice
	// Now the first column is "Index" but we will fill it with hex size values
	data := [][]string{
		{"Offset", "Field", "Value", "Size"},
	}

	var longestFieldName = 0
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		size := field.Type.Size()

		// Handle arrays separately
		var valueStr string
		if value.Kind() == reflect.Array {
			for j := 0; j < value.Len(); j++ {
				valueStr += fmt.Sprintf("%#x ", value.Index(j).Interface())
			}
		} else {
			valueStr = fmt.Sprintf("%#x", value.Interface())
		}

		// Instead of i, show size in hex as the "index"
		data = append(data, []string{
			fmt.Sprintf("0x%X", offset), // hex representation of size
			field.Name,
			valueStr,
			fmt.Sprintf("%d", size), // keep decimal bytes in the "Size" column
		})
		offset += size
		if len(field.Name) > longestFieldName {
			longestFieldName = len(field.Name)
		}
	}

	colWidths := []float32{65, float32(longestFieldName) * 10, 150, 100}
	colTypes := []ColumnType{hexCol, strCol, unsortableCol, decCol}
	return createNewSortableTable(colWidths, data, colTypes)
}

func createTableForDataDirectories(dataDirs []pe.DataDirectory, offset uintptr) (*sortableTable, error) {
	data := [][]string{
		{"Offset", "Directory", "RVA", "Size"},
	}

	var longestFieldName = 0

	for i, dir := range dataDirs {
		if i < len(directoryNames) {
			data = append(data, []string{fmt.Sprintf("0x%X", offset),
				directoryNames[i],
				fmt.Sprintf("0x%X", dir.VirtualAddress),
				fmt.Sprintf("%d", dir.Size)})

			if len(directoryNames[i]) > longestFieldName {
				longestFieldName = len(directoryNames[i])
			}
		}
		offset += 4
	}

	colWidths := []float32{65, float32(longestFieldName) * 10, 150, 100}
	colTypes := []ColumnType{hexCol, strCol, hexCol, decCol}
	return createNewSortableTable(colWidths, data, colTypes)
}

func createTableForSectionHeaders(sections []*pe.Section, offset uintptr) (*sortableTable, error) {

	data := [][]string{
		{"Offset", "Name", "Virtual Size", "Virtual Address",
			"Raw Size", "Raw data *", "Relocations *", "Relocations #",
			"Line Numbers *", "Line Numbers #", "Characteristics"},
	}

	for _, section := range sections {
		header := section.SectionHeader
		data = append(data, []string{
			fmt.Sprintf("0x%X", offset),
			header.Name,
			fmt.Sprintf("0x%X", header.VirtualSize),
			fmt.Sprintf("0x%X", header.VirtualAddress),
			fmt.Sprintf("%d", header.Size),
			fmt.Sprintf("0x%X", header.Offset),
			fmt.Sprintf("0x%X", header.PointerToRelocations),
			fmt.Sprintf("%d", header.NumberOfRelocations),
			fmt.Sprintf("0x%X", header.PointerToLineNumbers),
			fmt.Sprintf("%d", header.NumberOfLineNumbers),
			fmt.Sprintf("0x%X", header.Characteristics)})
		offset += 0x28
	}

	colWidths := []float32{65, 80, 100, 110, 100, 100, 100, 100, 120, 120, 110}
	colTypes := []ColumnType{hexCol, strCol, hexCol, hexCol, decCol, hexCol, hexCol, decCol,
		hexCol, decCol, hexCol}
	return createNewSortableTable(colWidths, data, colTypes)
}

func createNewSortableTable(colWidths []float32, data [][]string, colTypes []ColumnType) (*sortableTable, error) {

	// Measure row heights (assuming measureRowsHeights supports 4 columns)
	colHeights := measureRowsHeights(data, colWidths)

	st := newSortableTable(data, colWidths, colTypes)

	// Apply column widths
	for colIndex, width := range colWidths {
		st.table.SetColumnWidth(colIndex, width)
	}
	// Apply row heights
	for row, height := range colHeights {
		st.table.SetRowHeight(row, height)
	}
	return st, nil
}
