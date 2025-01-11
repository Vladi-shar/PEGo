package main

import (
	"bytes"
	"debug/pe"
	"fmt"
	"image/png"
	"os"
	"reflect"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	_ "embed"
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
			return widget.NewLabel("") // Template node
		},
		// Define how to update the template for a specific node
		func(uid widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(uid)
		},
	)

	ui.leftPane.Add(tree)
	var filePath = "C:\\Program Files\\Android\\Android Studio\\bin\\breakgen64.dll"
	// Create the File menu

	// fileMenu := fyne.NewMenu("File",
	// 	fyne.NewMenuItem("Open", func() {
	//var err error
	// filePath, err = dialog.File().Title("Select a File").Load()
	// if err != nil {
	// 	if err.Error() != "cancelled" { // Ignore "cancelled" error
	// 		fmt.Println("Error opening file:", err)
	// 	}
	// 	return
	// }

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

	peFull.dos = dos
	peFull.peFile = peFile

	// sections, _ := getSections(file)

	data = getPeTreeMap(peFile, filePath)

	// Update left and right panes (assuming `leftPane` and `rightPane` are defined widgets)
	// leftPane.SetText(sections)   // Set file name in left pane
	tree.Root = "File: " + filePath
	tree.Refresh()
	tree.OpenAllBranches()
	//}),
	//)

	tree.OnSelected = func(uid widget.TreeNodeID) {
		if uid == "Dos Header" {
			// Call the function to display DOS header details
			displayDosHeaderDetails(ui, peFull.dos)
		} else {
			ui.rightPane.RemoveAll()
			ui.rightPane.Add(widget.NewLabel(uid))
		}
	}

	// Create the main menu
	//mainMenu := fyne.NewMainMenu(fileMenu)

	// Create a horizontal split
	split := container.NewHSplit(ui.leftPane, ui.rightPane)
	split.SetOffset(0.25) // Set the split ratio (0.5 means equal halves)
	fixedSplit := container.NewStack(split)

	// Set the content with a vertical layout
	content := container.NewBorder(nil, nil, nil, nil, fixedSplit)

	// Set the menu and content in the window
	//window.SetMainMenu(mainMenu)
	window.SetContent(content)

	// Show and run the application
	window.Resize(fyne.NewSize(800, 600))
	window.ShowAndRun()
}

func createTableFromStruct(header any) (*widget.Table, error) {
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
	data := [][]string{{"Field", "Value", "Size"}} // Header row

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

		data = append(data, []string{field.Name, valueStr, fmt.Sprintf("%d", size)})
	}

	var table *widget.Table // declare a pointer to a Table

	// Create the table
	table = widget.NewTable(
		func() (int, int) {
			return len(data), len(data[0]) // Number of rows and columns
		},
		func() fyne.CanvasObject {
			lbl := widget.NewLabel("") // Template for each cell
			lbl.Wrapping = fyne.TextWrapWord
			return lbl
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			lbl := cell.(*widget.Label)
			text := data[id.Row][id.Col]
			lbl.SetText(text) // Populate cell content

			size := fyne.MeasureText(text, lbl.Size().Height, lbl.TextStyle)
			lines := float32(size.Width) / 200
			if lines < 1 {
				lines = 1
			}

			neededHeight := float32(size.Height) * fyne.Max(lines, 1)
			padding := float32(8)
			table.SetRowHeight(id.Row, neededHeight+padding)
			// table.SetRowHeight(id.Row, 10)
		},
	)

	return table, nil
}
