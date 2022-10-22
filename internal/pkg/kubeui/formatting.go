package kubeui

import (
	"fmt"
	"kubeui/internal/pkg/k8s"
	"sort"
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
	"github.com/life4/genesis/maps"
	"github.com/life4/genesis/slices"
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/integer"
)

// LineBreak splits a string into multiple lines to make each line a max of maxWidth in length.
func LineBreak(str string, maxWidth int) string {

	if len(str) <= maxWidth {
		return str
	}

	if len(strings.Split(str, " ")) == 1 && len(str) > maxWidth {
		return str[:maxWidth]
	}

	charCount := 0
	newStr := []rune{}

	var i int
	for i < len(str) {
		char := str[i]
		// If we have not yet reached the charCount.
		if charCount < maxWidth-1 {
			charCount++
			newStr = append(newStr, rune(char))
			i++
			continue
		}

		// If we have reached the charCount and we have a space, then we replace the space with a newline.
		if str[i] == ' ' {
			newStr = append(newStr, '\n')
			charCount = 0
			i++
			continue
		}

		newStr = append(newStr, '\n')
		newStr = append(newStr, rune(char))
		charCount = 0
		i++
	}
	return string(newStr)

}

// Truncate truncates a string and adds ... to indicate it was truncated.
// It keeps words intact as falls back to the last word with room for ...
func Truncate(text string, maxLen int) string {

	maxIdx := maxLen - 3

	if maxIdx <= 0 {
		return ""
	}

	lastSpace := maxIdx
	len := 0
	for i, r := range text {
		if unicode.IsSpace(r) {
			lastSpace = i
		}
		len++
		if len > maxIdx {
			return text[:lastSpace] + "..."
		}
	}

	return text
}

type DataColumn struct {
	desc  string
	width int
}

type DataRow struct {
	values []string
}

// EventsTable creates the neccessary columns and row in order to display event information.
func EventsTable(events []v1.Event) ([]*DataColumn, []*DataRow) {
	eventColumns := []*DataColumn{
		{desc: "Type", width: 4},
		{desc: "Reason", width: 6},
		{desc: "Age", width: 3},
		{desc: "From", width: 4},
		{desc: "Message", width: 50},
	}

	eventRows := slices.Map(events, func(e v1.Event) *DataRow {

		eventFormat := k8s.NewListEventFormat(e)

		// Update widths of the name and status columns
		eventColumns[0].width = integer.IntMax(eventColumns[0].width, len(eventFormat.Type))
		eventColumns[1].width = integer.IntMax(eventColumns[1].width, len(eventFormat.Reason))
		eventColumns[2].width = integer.IntMax(eventColumns[2].width, len(eventFormat.Age))
		eventColumns[3].width = integer.IntMax(eventColumns[3].width, len(eventFormat.From))
		eventColumns[4].width = integer.IntMin(eventColumns[4].width, len(eventFormat.Message))

		return &DataRow{
			values: []string{eventFormat.Type, eventFormat.Reason, eventFormat.Age, eventFormat.From, eventFormat.Message},
		}
	})

	return eventColumns, eventRows
}

// PodStatusTable creates the neccessary columns and row in order to display pod status information.
func PodStatusTable(pod v1.Pod) ([]*DataColumn, *DataRow) {
	podColumns := []*DataColumn{
		{desc: "Name", width: 4},
		{desc: "Ready", width: 5},
		{desc: "Status", width: 6},
		{desc: "Restarts", width: 8},
		{desc: "Age", width: 3},
	}

	// Update widths of the name and status columns

	podFormat := k8s.NewListPodFormat(pod)

	podColumns[0].width = integer.IntMax(podColumns[0].width, len(podFormat.Name))
	podColumns[1].width = integer.IntMax(podColumns[1].width, len(podFormat.Ready))
	podColumns[2].width = integer.IntMax(podColumns[2].width, len(podFormat.Status))
	podColumns[3].width = integer.IntMax(podColumns[3].width, len(podFormat.Restarts))
	podColumns[4].width = integer.IntMax(podColumns[4].width, len(podFormat.Age))

	podRow := &DataRow{
		values: []string{podFormat.Name, podFormat.Ready, podFormat.Status, podFormat.Restarts, podFormat.Age},
	}

	return podColumns, podRow
}

// StringMapTable creates the neccessary columns and rows in order to display a map[string]string as a table.
func StringMapTable(col1 string, col2 string, data map[string]string) ([]*DataColumn, []*DataRow) {

	columns := []*DataColumn{
		{desc: col1, width: len(col1)},
		{desc: col2, width: len(col2)},
	}

	keys := maps.Keys(data)
	sort.Strings(keys)

	rows := []*DataRow{}

	for _, key := range keys {

		value := data[key]
		// Update widths of the columns
		columns[0].width = integer.IntMax(columns[0].width, len(key))
		columns[1].width = integer.IntMax(columns[1].width, len(value))

		rows = append(rows, &DataRow{
			values: []string{key, value},
		})
	}

	return columns, rows
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
