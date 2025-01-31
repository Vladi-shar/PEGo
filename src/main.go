package main

import (
	"fyne.io/fyne/v2/app"
)

func main() {

	// Create the application
	MyApp = app.New()
	myWindow := MyApp.NewWindow("PEGo")

	icon := loadIcon()
	myWindow.SetIcon(icon)

	InitPaneView(myWindow)

}
