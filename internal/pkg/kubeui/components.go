package kubeui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/life4/genesis/slices"
	"github.com/muesli/reflow/wrap"
)

// ColumnTable converts a set of columns and rows to an aligned table string
func ColumnTable(columns []*DataColumn, rows []*DataRow) string {
	var builder strings.Builder

	builder.WriteString(ColumnsString(columns) + "\n\n")

	// Iterate over the rows in the current page and print them out.
	builder.WriteString(RowsString(columns, rows))

	return builder.String()
}

// ColumnsString converts a set of columns to an aligned string
func ColumnsString(columns []*DataColumn) string {

	columnsData := slices.Map(columns, func(c *DataColumn) string {
		return lipgloss.NewStyle().Width(c.width + 2).Render(c.desc)
	})

	return lipgloss.JoinHorizontal(lipgloss.Left, columnsData...)
}

// RowsString converts a set of rows to an aligned string
func RowsString(columns []*DataColumn, rows []*DataRow) string {

	var builder strings.Builder

	// Iterate over the rows in the current page and print them out.
	for _, row := range rows {

		rowData := []string{}

		for i, value := range row.values {
			rowData = append(rowData, lipgloss.NewStyle().Width(columns[i].width+2).Render(value))
		}
		// Render the row
		builder.WriteString(fmt.Sprintf("%s\n", lipgloss.JoinHorizontal(lipgloss.Left, rowData...)))
	}

	return builder.String()

}

// TabsSelect renders a tab select.
func TabsSelect(cursor, maxWidth int, headers []string) string {
	var tabsBuilder strings.Builder

	// Iterate over the items in the current page and print them out.
	for i, header := range headers {

		// Is the cursor pointing at this choice?
		if cursor == i {
			tabsBuilder.WriteString(lipgloss.NewStyle().Underline(true).Render(header) + " ")
			continue
		}

		tabsBuilder.WriteString(header + " ")
	}

	return lipgloss.NewStyle().Width(maxWidth).Render(tabsBuilder.String())
}

// HorizontalSelectList renders a horizontal list with an item highlighted as selected.
func HorizontalSelectList(items []string, selectedItem string, maxWidth int) string {

	builder := strings.Builder{}

	for i, item := range items {
		if item == selectedItem {
			builder.WriteString(SelectedStyle.Render(fmt.Sprintf("[%d] %s", i+1, item)) + " ")
			continue
		}
		builder.WriteString(UnselectedStyle.Render(fmt.Sprintf("[%d] %s", i+1, item)) + " ")
	}

	return wrap.String(builder.String(), maxWidth)
}

// StatusBar returns a string representing a status bar.
// Values are printed with separator in between.
func StatusBar(width int, separator string, values ...string) string {
	statusBar := StatusBarStyle.Width(width)

	builder := &strings.Builder{}

	for _, v := range values {
		builder.WriteString(fmt.Sprintf("%s%s", v, separator))
	}

	return statusBar.Render(lipgloss.NewStyle().Width(width).Render(strings.TrimRight(builder.String(), separator)))
}
