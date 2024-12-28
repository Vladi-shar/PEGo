package main

import (
	"fmt"

	"github.com/sqweek/dialog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	// Create the application
	myApp := app.New()
	myWindow := myApp.NewWindow("Fyne UI Example")

	// Create two panes
	leftPane := widget.NewLabel("Left Pane Content")
	rightPane := widget.NewLabel("Right Pane Content")

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

			fileData, err := readFile(filePath)
			if err != nil {
				fmt.Println("Failed to read file, error:", err)
				return
			}
			if !isPEFile(fileData) {
				fmt.Println("File is not a PE file")
				return
			}

			// Update the panes with file information
			sections, _ := getSections(fileData)
			dosHeader, _ := getDosHeader(fileData)

			// Update left and right panes (assuming `leftPane` and `rightPane` are defined widgets)
			leftPane.SetText(sections)   // Set file name in left pane
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
