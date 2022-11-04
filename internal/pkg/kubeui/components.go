package kubeui

import (
	"fmt"
	"kubeui/internal/pkg/kubeui/help"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/life4/genesis/slices"
)

// ShortHelp returns a short help view.
func ShortHelp(width int, keys []key.Binding) string {
	h := help.New()
	h.Width = width
	return h.ShortHelpView(keys)
}

// FullHelp returns a full help view.
func FullHelp(width int, keys [][]key.Binding) string {
	h := help.New()
	h.Width = width
	return h.FullHelpView(keys)
}

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
