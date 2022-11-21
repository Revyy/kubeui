package table

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/life4/genesis/slices"
)

// DataColumn represents a column in a table.
type DataColumn struct {
	Desc  string
	Width int
}

// DataRow represents a row in a table.
type DataRow struct {
	Values []string
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
		return lipgloss.NewStyle().Width(c.Width + 2).Render(c.Desc)
	})

	return lipgloss.JoinHorizontal(lipgloss.Left, columnsData...)
}

// RowsString converts a set of rows to an aligned string
func RowsString(columns []*DataColumn, rows []*DataRow) string {

	var builder strings.Builder

	// Iterate over the rows in the current page and print them out.
	for _, row := range rows {

		rowData := []string{}

		for i, value := range row.Values {
			rowData = append(rowData, lipgloss.NewStyle().Width(columns[i].Width+2).Render(value))
		}
		// Render the row
		builder.WriteString(fmt.Sprintf("%s\n", lipgloss.JoinHorizontal(lipgloss.Left, rowData...)))
	}

	return builder.String()

}
