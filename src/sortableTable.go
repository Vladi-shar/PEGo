package main

import (
	"fmt"
	"sort"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type selectableLabel struct {
	widget.Entry
	originalText string
}

// newSelectableLabel creates a new selectableLabel with the given text.
func newSelectableLabel(text string) *selectableLabel {
	s := &selectableLabel{originalText: text}
	// Set the text and the change-reversion hack.
	s.Entry.MultiLine = true
	s.Entry.SetText(text)
	s.Entry.OnChanged = func(newText string) {
		if newText != s.originalText {
			s.Entry.SetText(s.originalText)
		}
	}
	s.Entry.Wrapping = fyne.TextWrapWord
	// Register the composite widget.
	s.ExtendBaseWidget(s)
	return s
}

// CreateRenderer overrides the renderer to remove the default border and background.
func (s *selectableLabel) CreateRenderer() fyne.WidgetRenderer {
	// Get the original renderer.
	renderer := s.Entry.CreateRenderer()
	// Get the desired background color from the theme.
	bgColor := theme.Color(theme.ColorNameBackground)
	// Modify any canvas.Rectangle in the renderer: remove its stroke and force its fill.
	for _, obj := range renderer.Objects() {
		if rect, ok := obj.(*canvas.Rectangle); ok {
			rect.StrokeWidth = 0
			rect.FillColor = bgColor
			rect.StrokeColor = bgColor
		}
	}
	return renderer
}

// columnType is an enum for the type of data in a column.
type ColumnType int

const (
	hexCol ColumnType = iota
	decCol
	strCol
	unsortableCol
)

type ColumnProps struct {
	sortable   bool
	selectable bool
}

// sortableTable wraps a widget.Table plus the underlying data slice.
// It handles sorting when the user clicks a column header.
type sortableTable struct {
	table     *widget.Table
	data      [][]string
	colWidths []float32
	colTypes  []ColumnType
	colProps  []ColumnProps
	// Track the current sort direction per column (true=asc, false=desc)
	sortAsc map[int]bool
}

// newSortableTable creates a new sortableTable around an existing data set.
func newSortableTable(data [][]string, colWidths []float32, colTypes []ColumnType, colProps []ColumnProps) *sortableTable {
	st := &sortableTable{
		data:      data,
		colWidths: colWidths,
		sortAsc:   make(map[int]bool),
		colTypes:  colTypes,
		colProps:  colProps,
	}

	tbl := widget.NewTable(
		// Number of rows and columns:
		func() (int, int) {
			return len(st.data), len(st.data[0])
		},
		// Create: return a container with a background rect + label (default)
		func() fyne.CanvasObject {
			rect := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))

			lbl := widget.NewLabel("")
			lbl.Wrapping = fyne.TextWrapWord
			return container.NewStack(rect, lbl)
		},
		// Update: set the text, and for row=0 create a clickable "headerLabel".
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			c := obj.(*fyne.Container)
			rect := c.Objects[0].(*canvas.Rectangle)
			text := st.data[id.Row][id.Col]

			if id.Row == 0 {
				// This is the header row
				rect.FillColor = theme.Color(theme.ColorNameInputBackground)

				// Create a clickable headerLabel that, when tapped,
				// sorts the table by this column.
				clickable := newHeaderLabel(widget.NewLabel(text), func() {
					st.sortByColumn(id.Col)
				})
				c.Objects = []fyne.CanvasObject{rect, clickable}
			} else {
				var entry fyne.CanvasObject
				if st.colProps[id.Col].selectable {
					entry = newSelectableLabel(text)
				} else {
					// Just a normal label with wrapping
					entry = widget.NewLabel(text)
					entry.(*widget.Label).Wrapping = fyne.TextWrapWord
				}
				c.Objects = []fyne.CanvasObject{rect, entry}

			}
		},
	)

	st.table = tbl
	st.updateRowHeights()
	return st
}

// sortByColumn sorts st.data (excluding row 0, which is the header) by the given col index.
func (st *sortableTable) sortByColumn(col int) {
	// If this column is "unsortable", just return (do nothing)
	if !st.colProps[col].sortable {
		return
	}
	// Toggle the sort direction for this column
	st.sortAsc[col] = !st.sortAsc[col]
	ascending := st.sortAsc[col]

	// Sort the data in place, skipping row 0 (the header)
	sort.Slice(st.data[1:], func(i, j int) bool {
		leftStr := st.data[1+i][col]
		rightStr := st.data[1+j][col]

		switch st.colTypes[col] {
		case hexCol:
			leftVal := parseHex(leftStr)
			rightVal := parseHex(rightStr)
			if ascending {
				return leftVal < rightVal
			}
			return leftVal > rightVal

		case strCol:
			if ascending {
				return leftStr < rightStr
			}
			return leftStr > rightStr

		case decCol:
			leftVal := parseDec(leftStr)
			rightVal := parseDec(rightStr)
			if ascending {
				return leftVal < rightVal
			}
			return leftVal > rightVal
		}

		// Fallback (in case we add columns later):
		return false
	})

	// Re-measure row heights in case anything changed
	st.updateRowHeights()

	// Refresh the Table to see the changes
	st.table.Refresh()
}

// parseHex attempts to parse a string like "0x10" or "0XFF" into an int64.
// It automatically handles 0x prefix if you pass base=0 to ParseInt().
func parseHex(s string) int64 {
	val, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		fmt.Println("parseHex error:", err)
		return 0
	}
	return val
}

// parseDec attempts to parse a decimal string like "16" into an int64.
func parseDec(s string) int64 {
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		fmt.Println("parseDec error:", err)
		return 0
	}
	return val
}

// updateRowHeights measures each row’s height and calls SetRowHeight.
// If you’re using wrapping or variable text, you’ll want to measure it exactly.
// For simplicity, here we just call a measureRowsHeights function (shown below).
func (st *sortableTable) updateRowHeights() {
	rowHeights := measureRowsHeights(st.data, st.colWidths)
	for row, height := range rowHeights {
		st.table.SetRowHeight(row, height)
	}
}

func (st *sortableTable) removeRow(row int) {
	if row < 0 || row >= len(st.data) {
		// Index out of bounds, return unchanged
		return
	}

	st.data = append(st.data[:row], st.data[row+1:]...)
	st.updateRowHeights()
}

// newHeaderLabel wraps a standard label in a Tappable so we can detect clicks on it.
func newHeaderLabel(l *widget.Label, tapped func()) fyne.CanvasObject {
	btn := &headerLabel{
		Label:  l,
		tapped: tapped,
	}
	// We embed the label in a BaseWidget so it can receive events
	btn.ExtendBaseWidget(btn)
	return btn
}

// headerLabel is a clickable container for a label in the table header.
type headerLabel struct {
	*widget.Label
	tapped func()
}

// MinSize / layout / etc. are inherited from the base label, so no custom layout is needed.

// Tapped is called by the Fyne event system when the user clicks the header cell.
func (h *headerLabel) Tapped(_ *fyne.PointEvent) {
	if h.tapped != nil {
		h.tapped()
	}
}

// TappedSecondary is required for fyne.Tappable but not used here.
func (h *headerLabel) TappedSecondary(_ *fyne.PointEvent) {}

// measureRowsHeights is the same logic you had before: measure each cell's text
// *with wrapping* at the specified column widths, find the max height per row, etc.
func measureRowsHeights(data [][]string, colWidths []float32) map[int]float32 {
	rowHeights := make(map[int]float32)
	for rowIndex, row := range data {
		var maxHeight float32
		for colIndex, text := range row {
			// Create a wrapping label to measure
			lbl := widget.NewLabel(text)
			lbl.Wrapping = fyne.TextWrapWord

			// Force the label to measure at the desired column width
			desiredWidth := colWidths[colIndex]

			// We'll put it in a container so it can expand vertically
			c := container.NewWithoutLayout(lbl)
			c.Resize(fyne.NewSize(desiredWidth, 2000)) // plenty of height
			lbl.Resize(fyne.NewSize(desiredWidth, 2000))
			lbl.Refresh()

			sz := lbl.MinSize()
			if sz.Height > maxHeight {
				maxHeight = sz.Height
			}
		}
		rowHeights[rowIndex] = maxHeight
	}
	return rowHeights
}
