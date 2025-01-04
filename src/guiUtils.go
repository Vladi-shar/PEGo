package main

import (
	"bytes"
	"debug/pe"
	"fmt"
	"image/png"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/sqweek/dialog"

	_ "embed"
)

//go:embed winres\\logosmall.png
var iconBytes []byte

var MyApp fyne.App

type MyAppUI struct {
	leftPane  *fyne.Container
	rightPane *widget.Label
}

func initUIElements() *MyAppUI {
	return &MyAppUI{
		leftPane:  container.NewStack(),
		rightPane: widget.NewLabel("Right Pane Content"),
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
	ui.rightPane.SetText(msg)
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
				errorMessage := fmt.Sprintf("Error opening file: %v", err)
				displayErrorOnRightPane(ui, "Unsupported file format")
				fmt.Println(errorMessage)
				return
			}
			defer file.Close()

			// sections, _ := getSections(file)
			dosHeader, _ := getDosHeader(file)


			data = getPeTreeMap(file, filePath)

			// Update left and right panes (assuming `leftPane` and `rightPane` are defined widgets)
			// leftPane.SetText(sections)   // Set file name in left pane
			tree.Root = "File: " + filePath
			tree.Refresh()
			tree.OpenAllBranches()
			ui.rightPane.SetText(dosHeader) // Set file content in right pane
		}),
	)

	// Create the main menu
	mainMenu := fyne.NewMainMenu(fileMenu)

	// Create a horizontal split
	split := container.NewHSplit(ui.leftPane, ui.rightPane)
	split.SetOffset(0.25) // Set the split ratio (0.5 means equal halves)
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
