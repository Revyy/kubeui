package pods

import (
	"github.com/life4/genesis/slices"
	v1 "k8s.io/api/core/v1"
)

// Pod contains extended information about a kubernetes pod.
type Pod struct {
	Pod    v1.Pod
	Events []v1.Event

	// Stores the latest logs for each container, mapped on the container name.
	Logs map[string]string
}

// ContainerNames returns the names of all containers in the pod.
func (p *Pod) ContainerNames() []string {

	if p == nil {
		return []string{}
	}

	return slices.Map(p.Pod.Status.ContainerStatuses, func(s v1.ContainerStatus) string {
		return s.Name
	})
}
