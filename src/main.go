package main

import (
	"fmt"

	"debug/pe"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/sqweek/dialog"
)

func getSections(peFile *pe.File) (string, error) {

	return "not implemented", nil
}

func main() {
	// Create the application
	myApp := app.New()
	myWindow := myApp.NewWindow("Fyne UI Example")

	// Create two panes
	leftPane := container.NewStack()
	rightPane := widget.NewLabel("Right Pane Content")
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

	leftPane.Add(tree)

	// Create the File menu
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Open", func() {
			filePath, err := dialog.File().Title("Select a File").Load()
			if err != nil {
				if err.Error() != "cancelled" { // Ignore "cancelled" error
					fmt.Println("Error opening file:", err)
				}
				return
			}

			file, err := pe.Open(filePath)
			if err != nil {
				fmt.Println("Error opening file:", err)
				return
			}
			defer file.Close()

			// sections, _ := getSections(file)
			dosHeader, _ := getDosHeader(file)

			// Define the data for the tree
			// Simulate parsing the file and updating the data map
			// Replace this hardcoded data with your logic for populating data
			data[""] = []string{"File: " + filePath}
			data["File: "+filePath] = []string{"Dos Header", "Nt Headers", "Section Headers"}
			data["Nt Headers"] = []string{"File Header", "Optional Header"}
			data["Optional Header"] = []string{"Data Directories [x]"}
			data["Section Headers"] = []string{"Export Directory", "Import Directory", "Debug Directory"}

			// Update left and right panes (assuming `leftPane` and `rightPane` are defined widgets)
			// leftPane.SetText(sections)   // Set file name in left pane
			tree.Refresh()
			tree.OpenAllBranches()
			rightPane.SetText(dosHeader) // Set file content in right pane
		}),
	)

	// Create the main menu
	mainMenu := fyne.NewMainMenu(fileMenu)

	// Create a horizontal split
	split := container.NewHSplit(leftPane, rightPane)
	split.SetOffset(0.25) // Set the split ratio (0.5 means equal halves)
	fixedSplit := container.NewStack(split)

	// Set the content with a vertical layout
	content := container.NewBorder(nil, nil, nil, nil, fixedSplit)

	// Set the menu and content in the window
	myWindow.SetMainMenu(mainMenu)
	myWindow.SetContent(content)

	// Show and run the application
	myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.ShowAndRun()
}
