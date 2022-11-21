package podinfo

import (
	"kubeui/internal/pkg/k8s"
	"kubeui/internal/pkg/ui/table"

	"github.com/life4/genesis/slices"
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/integer"
)

// EventColumnsAndRows creates the neccessary columns and row in order to display event information.
func EventColumnsAndRows(maxWidth int, events []v1.Event) ([]*table.DataColumn, []*table.DataRow) {
	eventColumns := []*table.DataColumn{
		{Desc: "Type", Width: 4},
		{Desc: "Reason", Width: 6},
		{Desc: "Age", Width: 3},
		{Desc: "From", Width: 4},
		{Desc: "Message", Width: 50},
	}

	eventRows := slices.Map(events, func(e v1.Event) *table.DataRow {

		eventFormat := k8s.NewListEventFormat(e)

		// Update widths of the name and status columns
		eventColumns[0].Width = integer.IntMax(eventColumns[0].Width, len(eventFormat.Type))
		eventColumns[1].Width = integer.IntMax(eventColumns[1].Width, len(eventFormat.Reason))
		eventColumns[2].Width = integer.IntMax(eventColumns[2].Width, len(eventFormat.Age))
		eventColumns[3].Width = integer.IntMax(eventColumns[3].Width, len(eventFormat.From))

		remainingWidth := maxWidth - slices.Reduce(eventColumns[0:5], 0, func(c *table.DataColumn, acc int) int {
			return acc + c.Width
		})

		eventColumns[4].Width = integer.IntMax(remainingWidth-1, len(eventFormat.Message))

		if eventColumns[4].Width < 30 {
			eventColumns[4].Width = 30
		}

		return &table.DataRow{
			Values: []string{eventFormat.Type, eventFormat.Reason, eventFormat.Age, eventFormat.From, eventFormat.Message},
		}
	})

	return eventColumns, eventRows
}

// PodStatusColumnsAndRows creates the neccessary columns and row in order to display pod status information.
func PodStatusColumnsAndRows(pod v1.Pod) ([]*table.DataColumn, *table.DataRow) {
	podColumns := []*table.DataColumn{
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

	podRow := &table.DataRow{
		Values: []string{podFormat.Name, podFormat.Ready, podFormat.Status, podFormat.Restarts, podFormat.Age},
	}

	return podColumns, podRow
}
