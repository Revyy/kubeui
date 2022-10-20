package k8s

import v1 "k8s.io/api/core/v1"

// Pod contains extended information about a kubernetes pod.
type Pod struct {
	Pod    v1.Pod
	Events []v1.Event
}
