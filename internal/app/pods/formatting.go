package pods

import (
	"fmt"
	"kubeui/internal/pkg/component/columntable"
	"strconv"

	"github.com/life4/genesis/slices"
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/integer"
)

func podTableContents(pods []v1.Pod) ([]*columntable.Column, []*columntable.Row) {
	podColumns := []*columntable.Column{
		{Desc: "Name", Width: 30},
		{Desc: "Ready", Width: 5},
		{Desc: "Status", Width: 6},
		{Desc: "Restarts", Width: 8},
	}

	podRows := slices.Map(pods, func(p v1.Pod) *columntable.Row {

		readyCount := 0
		for _, status := range p.Status.ContainerStatuses {
			if status.Ready {
				readyCount++
			}
		}

		maxRestarts := 0
		for _, c := range p.Status.ContainerStatuses {
			maxRestarts = integer.IntMax(maxRestarts, int(c.RestartCount))
		}

		// Update widths of the name and status columns
		podColumns[0].Width = integer.IntMax(podColumns[0].Width, len(p.Name))
		podColumns[2].Width = integer.IntMax(podColumns[2].Width, len(p.Status.Phase))

		return &columntable.Row{
			Id:     p.Name,
			Values: []string{p.Name, fmt.Sprintf("%d/%d", readyCount, len(p.Status.ContainerStatuses)), string(p.Status.Phase), strconv.Itoa(maxRestarts)},
		}
	})

	return podColumns, podRows
}
