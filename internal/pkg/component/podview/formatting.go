package podview

import (
	"fmt"
	"kubeui/internal/pkg/component/columntable"
	"kubeui/internal/pkg/k8s"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/life4/genesis/maps"
	"github.com/life4/genesis/slices"
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/integer"
)

// eventsTable creates the neccessary columns and row in order to display event information.
func eventsTable(events []v1.Event) ([]*columntable.Column, []*columntable.Row) {
	eventColumns := []*columntable.Column{
		{Desc: "Type", Width: 4},
		{Desc: "Reason", Width: 6},
		{Desc: "Age", Width: 3},
		{Desc: "From", Width: 4},
		{Desc: "Message", Width: 50},
	}

	eventRows := slices.Map(events, func(e v1.Event) *columntable.Row {

		eventFormat := k8s.NewListEventFormat(e)

		// Update widths of the name and status columns
		eventColumns[0].Width = integer.IntMax(eventColumns[0].Width, len(eventFormat.Type))
		eventColumns[1].Width = integer.IntMax(eventColumns[1].Width, len(eventFormat.Reason))
		eventColumns[2].Width = integer.IntMax(eventColumns[2].Width, len(eventFormat.Age))
		eventColumns[3].Width = integer.IntMax(eventColumns[3].Width, len(eventFormat.From))
		eventColumns[4].Width = integer.IntMin(eventColumns[4].Width, len(eventFormat.Message))

		return &columntable.Row{
			Id:     e.InvolvedObject.Name,
			Values: []string{eventFormat.Type, eventFormat.Reason, eventFormat.Age, eventFormat.From, eventFormat.Message},
		}
	})

	return eventColumns, eventRows
}

// podStatusTable creates the neccessary columns and row in order to display pod information.
func podStatusTable(pod v1.Pod) ([]*columntable.Column, *columntable.Row) {
	podColumns := []*columntable.Column{
		{Desc: "Name", Width: 4},
		{Desc: "Ready", Width: 5},
		{Desc: "Status", Width: 6},
		{Desc: "Restarts", Width: 8},
		{Desc: "Age", Width: 3},
	}

	// Update widths of the name and status columns

	podFormat := k8s.NewListPodFormat(pod)

	podColumns[0].Width = integer.IntMax(podColumns[0].Width, len(podFormat.Name))
	podColumns[1].Width = integer.IntMax(podColumns[1].Width, len(podFormat.Ready))
	podColumns[2].Width = integer.IntMax(podColumns[2].Width, len(podFormat.Status))
	podColumns[3].Width = integer.IntMax(podColumns[3].Width, len(podFormat.Restarts))
	podColumns[4].Width = integer.IntMax(podColumns[4].Width, len(podFormat.Age))

	podRow := &columntable.Row{
		Id:     pod.Name,
		Values: []string{podFormat.Name, podFormat.Ready, podFormat.Status, podFormat.Restarts, podFormat.Age},
	}

	return podColumns, podRow
}

// stringMapTable creates the neccessary columns and rows in order to display pod annotations.
func stringMapTable(col1 string, col2 string, data map[string]string) ([]*columntable.Column, []*columntable.Row) {

	columns := []*columntable.Column{
		{Desc: col1, Width: len(col1)},
		{Desc: col2, Width: len(col2)},
	}

	keys := maps.Keys(data)
	sort.Strings(keys)

	rows := []*columntable.Row{}

	for _, key := range keys {

		value := data[key]
		// Update widths of the columns
		columns[0].Width = integer.IntMax(columns[0].Width, len(key))
		columns[1].Width = integer.IntMax(columns[1].Width, len(value))

		rows = append(rows, &columntable.Row{
			Id:     key,
			Values: []string{key, value},
		})
	}

	return columns, rows
}

// columnTableData converts a set of columns and rows to an aligned table string
func columnTableData(columns []*columntable.Column, rows []*columntable.Row) string {
	var builder strings.Builder

	builder.WriteString(columnsString(columns) + "\n\n")

	// Iterate over the rows in the current page and print them out.
	builder.WriteString(rowsString(columns, rows))

	return builder.String()
}

// columnsString converts a set of columns to an aligned string
func columnsString(columns []*columntable.Column) string {

	columnsData := slices.Map(columns, func(c *columntable.Column) string {
		return lipgloss.NewStyle().Width(c.Width + 2).Render(c.Desc)
	})

	return lipgloss.JoinHorizontal(lipgloss.Left, columnsData...)
}

// rowsString converts a set of rows to an aligned string
func rowsString(columns []*columntable.Column, rows []*columntable.Row) string {

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

// tabsBuilder renders a tab select.
func tabsBuilder(cursor, maxWidth int, headers []string) string {
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
