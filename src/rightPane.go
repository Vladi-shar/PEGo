package main

func displayDosHeaderDetails(ui *MyAppUI, dosHeader *DOSHeader) {

	table, err := createTableFromStruct(dosHeader)
	if err != nil {
		displayErrorOnRightPane(ui, err.Error())
		return
	}

	// Replace rightPane with the table
	ui.rightPane.RemoveAll()
	ui.rightPane.Add(table)

}
