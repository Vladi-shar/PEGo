package main

import (
	"bytes"
	"debug/pe"
	_ "embed"
	"fmt"
	"image/png"
	"os"
	"reflect"
	"unsafe"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
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

	var peFull PE_FULL

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

			peFull.dos = dos
			peFull.nt = nt
			peFull.peFile = peFile

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
			fmt.Printf("sizeof nt: %d\n", unsafe.Sizeof(peFull.nt))
			fmt.Printf("sizeof nt.signature: %d\n", unsafe.Sizeof(peFull.nt.Signature))
			dummy := dummy{}
			fmt.Printf("sizeof dummy: %d\n", unsafe.Sizeof(dummy))
			displayFileHeaderDetails(ui, &peFull.peFile.FileHeader, uintptr(peFull.dos.E_ifanew)+unsafe.Sizeof(peFull.nt))
		case "Optional Header":
			optHeader, err := getOptionalHeader(peFull.peFile)
			if err != nil {
				displayErrorOnRightPane(ui, err.Error())
				return
			}
			displayOptionalHeaderDetails(ui, optHeader, uintptr(peFull.dos.E_ifanew)+unsafe.Sizeof(peFull.nt)+unsafe.Sizeof(peFull.peFile.FileHeader))
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

func getLongestString(strings []string) string {
	var longest string
	for _, str := range strings {
		if len(str) > len(longest) {
			longest = str
		}
	}
	return longest
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

	fmt.Println("Longest field name:", longestFieldName)
	// Now we have 4 columns, so adjust widths accordingly

	// get the current text size
	textSize := theme.TextSize()
	fmt.Println("Text size:", textSize)

	colWidths := []float32{65, float32(longestFieldName) * 10, 150, 100}

	// Measure row heights (assuming measureRowsHeights supports 4 columns)
	colHeights := measureRowsHeights(data, colWidths)

	// // Create the table
	// table := widget.NewTable(
	// 	func() (int, int) {
	// 		return len(data), len(data[0]) // Number of rows/columns
	// 	},
	// 	func() fyne.CanvasObject {
	// 		lbl := widget.NewLabel("") // Template for each cell
	// 		lbl.Wrapping = fyne.TextWrapWord
	// 		return lbl
	// 	},
	// 	func(id widget.TableCellID, cell fyne.CanvasObject) {
	// 		cell.(*widget.Label).SetText(data[id.Row][id.Col])
	// 	},
	// )

	st := newSortableTable(data, colWidths)

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
