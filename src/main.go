package main

import (
	"fmt"

	"github.com/sqweek/dialog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	// "fyne.io/fyne/v2/dialog"
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
		fyne.NewMenuItem("New", func() {
			// Action for New
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Menu Action",
				Content: "New File action triggered",
			})
		}),
		fyne.NewMenuItem("Open", func() {
			filePath, err := dialog.File().Title("Select a File").Load()
			if err != nil {
				if err.Error() != "cancelled" { // Ignore "cancelled" error
					fmt.Println("Error opening file:", err)
				}
				return
			}

			// // Function to get the file name
			// getFileName := func(filePath string) string {
			// 	return filepath.Base(filePath)
			// }

			// // Function to get the file contents
			// getFileContents := func(filePath string) (string, error) {
			// 	file, err := os.Open(filePath)
			// 	if err != nil {
			// 		return "", err
			// 	}
			// 	defer file.Close()

			// 	content, err := io.ReadAll(file)
			// 	if err != nil {
			// 		return "", err
			// 	}
			// 	return string(content), nil
			// }

			// Update the panes with file information
			fileName := getFileName(filePath)
			fileContents, err := getFileContents(filePath)
			if err != nil {
				fmt.Println("Error reading file contents:", err)
				return
			}

			// Update left and right panes (assuming `leftPane` and `rightPane` are defined widgets)
			leftPane.SetText(fileName)      // Set file name in left pane
			rightPane.SetText(fileContents) // Set file content in right pane
		}),
		fyne.NewMenuItem("Save", func() {
			// Action for Save
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Exit", func() {
			myApp.Quit()
		}),
	)

	// Create the main menu
	mainMenu := fyne.NewMainMenu(fileMenu)

	// Create a horizontal split
	split := container.NewHSplit(leftPane, rightPane)
	split.SetOffset(0.5) // Set the split ratio (0.5 means equal halves)

	// Set the content with a vertical layout
	content := container.NewBorder(nil, nil, nil, nil, split)

	// Set the menu and content in the window
	myWindow.SetMainMenu(mainMenu)
	myWindow.SetContent(content)

	// Show and run the application
	myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.ShowAndRun()
}
