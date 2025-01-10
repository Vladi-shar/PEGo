package main

func displayDosHeaderDetails(ui *MyAppUI, dosHeader *DOSHeader) {

	table, err := createTableFromStruct(dosHeader)
	if err != nil {
		displayErrorOnRightPane(ui, err.Error())
		return
	}

	table.SetColumnWidth(0, 150) // Width for the "Field" column
	table.SetColumnWidth(1, 200) // Width for the "Value" column

	// Replace rightPane with the table
	ui.rightPane.RemoveAll()
	ui.rightPane.Add(table)

}
