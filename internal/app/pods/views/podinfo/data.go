package podinfo

import (
	"kubeui/internal/pkg/k8s"
	"kubeui/internal/pkg/ui/table"
	"sort"

	"github.com/life4/genesis/maps"
	"github.com/life4/genesis/slices"
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/integer"
)

// eventColumnsAndRows creates the neccessary columns and row in order to display event information.
func eventColumnsAndRows(maxWidth int, events []v1.Event) ([]table.DataColumn, []table.DataRow) {
	eventColumns := []table.DataColumn{
		{Desc: "Type", Width: 6},
		{Desc: "Reason", Width: 8},
		{Desc: "Age", Width: 5},
		{Desc: "From", Width: 6},
		{Desc: "Message", Width: 50},
	}

	eventRows := slices.Map(events, func(e v1.Event) table.DataRow {

		eventFormat := k8s.NewListEventFormat(e)

		// Update widths of the name and status columns
		eventColumns[0].Width = integer.IntMax(eventColumns[0].Width, len(eventFormat.Type)+2)
		eventColumns[1].Width = integer.IntMax(eventColumns[1].Width, len(eventFormat.Reason)+2)
		eventColumns[2].Width = integer.IntMax(eventColumns[2].Width, len(eventFormat.Age)+2)
		eventColumns[3].Width = integer.IntMax(eventColumns[3].Width, len(eventFormat.From))

		remainingWidth := maxWidth - slices.Reduce(eventColumns[0:5], 0, func(c table.DataColumn, acc int) int {
			return acc + c.Width
		})

		eventColumns[4].Width = integer.IntMax(remainingWidth-1, len(eventFormat.Message))

		if eventColumns[4].Width < 30 {
			eventColumns[4].Width = 30
		}

		return table.DataRow{
			Values: []string{eventFormat.Type, eventFormat.Reason, eventFormat.Age, eventFormat.From, eventFormat.Message},
		}
	})

	return eventColumns, eventRows
}

// podStatusColumnsAndRows creates the neccessary columns and row in order to display pod status information.
func podStatusColumnsAndRows(pod v1.Pod) ([]table.DataColumn, table.DataRow) {
	podColumns := []table.DataColumn{
		{Desc: "Name", Width: 6},
		{Desc: "Ready", Width: 7},
		{Desc: "Status", Width: 8},
		{Desc: "Restarts", Width: 10},
		{Desc: "Age", Width: 3},
	}

	// Update widths of the name and status columns

	podFormat := k8s.NewListPodFormat(pod)

	podColumns[0].Width = integer.IntMax(podColumns[0].Width, len(podFormat.Name)+2)
	podColumns[1].Width = integer.IntMax(podColumns[1].Width, len(podFormat.Ready)+2)
	podColumns[2].Width = integer.IntMax(podColumns[2].Width, len(podFormat.Status)+2)
	podColumns[3].Width = integer.IntMax(podColumns[3].Width, len(podFormat.Restarts)+2)
	podColumns[4].Width = integer.IntMax(podColumns[4].Width, len(podFormat.Age))

	podRow := table.DataRow{
		Values: []string{podFormat.Name, podFormat.Ready, podFormat.Status, podFormat.Restarts, podFormat.Age},
	}

	return podColumns, podRow
}

// stringMapColumnsAndRows creates the neccessary columns and rows in order to display a map[string]string as a table.
func stringMapColumnsAndRows(maxWidth int, col1 string, col2 string, data map[string]string) ([]table.DataColumn, []table.DataRow) {

	columns := []table.DataColumn{
		{Desc: col1, Width: len(col1) + 2},
		{Desc: col2, Width: len(col2)},
	}

	keys := maps.Keys(data)
	sort.Strings(keys)

	rows := []table.DataRow{}

	for _, key := range keys {

		value := data[key]
		// Update widths of the columns
		columns[0].Width = integer.IntMax(columns[0].Width, len(key)+2)

		remainingWidth := maxWidth - columns[0].Width
		columns[1].Width = integer.IntMax(remainingWidth-1, len(value))

		rows = append(rows, table.DataRow{
			Values: []string{key, value},
		})
	}

	return columns, rows
}
