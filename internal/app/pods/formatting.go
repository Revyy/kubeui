package pods

import (
	"kubeui/internal/pkg/component/columntable"
	"kubeui/internal/pkg/k8s"

	"github.com/life4/genesis/slices"
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/integer"
)

func podTableContents(pods []v1.Pod) ([]*columntable.Column, []*columntable.Row) {
	podColumns := []*columntable.Column{
		{Desc: "Name", Width: 4},
		{Desc: "Ready", Width: 5},
		{Desc: "Status", Width: 6},
		{Desc: "Restarts", Width: 8},
		{Desc: "Age", Width: 3},
	}

	podRows := slices.Map(pods, func(p v1.Pod) *columntable.Row {

		podFormat := k8s.NewListPodFormat(p)

		// Update widths of the name and status columns
		podColumns[0].Width = integer.IntMax(podColumns[0].Width, len(p.Name))
		podColumns[1].Width = integer.IntMax(podColumns[1].Width, len(podFormat.Ready))
		podColumns[2].Width = integer.IntMax(podColumns[2].Width, len(podFormat.Status))
		podColumns[3].Width = integer.IntMax(podColumns[3].Width, len(podFormat.Restarts))
		podColumns[4].Width = integer.IntMax(podColumns[4].Width, len(podFormat.Age))

		return &columntable.Row{
			Id:     p.Name,
			Values: []string{podFormat.Name, podFormat.Ready, podFormat.Status, podFormat.Restarts, podFormat.Age},
		}
	})

	return podColumns, podRows
}
