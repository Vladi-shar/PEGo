package main

import (
	"debug/pe"

	"fyne.io/fyne/v2/app"
)

func getSections(peFile *pe.File) (string, error) {

	return "not implemented", nil
}

func main() {

	// Create the application
	MyApp = app.New()
	myWindow := MyApp.NewWindow("PEGo")

	icon := loadIcon()
	myWindow.SetIcon(icon)

	InitPaneView(myWindow)

}
