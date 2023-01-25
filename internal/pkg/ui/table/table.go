// Package table provides a way to render table like content as a single string.
// DataColumn describes the width of each column in the table along with a header, and the values defined in a DataRow are made to fit the width of each respective column.
package table

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"k8s.io/utils/integer"
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

// ColumnTable converts a set of columns and rows to an aligned table string.
func ColumnTable(columns []DataColumn, rows []DataRow) string {
	var builder strings.Builder

	builder.WriteString(ColumnsToString(columns) + "\n\n")

	// Iterate over the rows in the current page and print them out.
	builder.WriteString(RowsToString(columns, rows))

	return builder.String()
}

// ColumnsToString converts a set of columns to an aligned string.
// Column widths are expected to be equal or larger than the length of the descriptions, if not then the descriptions will be truncated.
func ColumnsToString(columns []DataColumn) string {

	columnsData := []string{}

	for _, c := range columns {
		columnsData = append(columnsData, lipgloss.NewStyle().Width(c.Width).Render(c.Desc[0:integer.IntMin(c.Width, len(c.Desc))]))
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, columnsData...)
}

// RowsToString converts a set of rows to an aligned string.
func RowsToString(columns []DataColumn, rows []DataRow) string {

	if len(columns) == 0 || len(rows) == 0 {
		return ""
	}

	rowStrings := []string{}

	// Iterate over the rows in the current page and print them out.
	for _, row := range rows {

		if len(row.Values) != len(columns) {
			continue
		}

		rowData := []string{}

		for i, value := range row.Values {
			rowData = append(rowData, lipgloss.NewStyle().Width(columns[i].Width).Render(value))
		}
		// Render the row
		rowStrings = append(rowStrings, lipgloss.JoinHorizontal(lipgloss.Left, rowData...))
	}

	return strings.Join(rowStrings, "\n")

}
