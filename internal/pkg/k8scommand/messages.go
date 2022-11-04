package k8scommand

import (
	"kubeui/internal/pkg/k8s"

	v1 "k8s.io/api/core/v1"
)

// ListNamespacesMsg is sent after fetching a list of available namespaces.
type ListNamespacesMsg struct {
	NamespaceList *v1.NamespaceList
}

// NewListNamespacesMsg creates a new ListNamespacesMsg message.
func NewListNamespacesMsg(namespaceList *v1.NamespaceList) ListNamespacesMsg {
	return ListNamespacesMsg{NamespaceList: namespaceList}
}

// ListPodsMsg is used as the result of fetching a list of pods in the current namespace.
type ListPodsMsg struct {
	PodList *v1.PodList
}

// NewListPodsMsg creates a new ListPods message.
func NewListPodsMsg(podList *v1.PodList) ListPodsMsg {
	return ListPodsMsg{PodList: podList}
}

// PodDeletedMsg is sent after a pod has been deleted.
type PodDeletedMsg struct {
	Name string
}

// PodDeletedMsg creates a new PodDeleted message.
func NewPodDeletedMsg(name string) PodDeletedMsg {
	return PodDeletedMsg{Name: name}
}

// GetPodMsg is used as the result of fetching a pod in the current namespace.
type GetPodMsg struct {
	Pod *k8s.Pod
}

// NewGetPod creates a new GetPod message.
func NewGetPodMsg(pod *v1.Pod, events []v1.Event, logs string) GetPodMsg {
	return GetPodMsg{Pod: &k8s.Pod{
		Pod:    *pod,
		Events: events,
		Logs:   logs,
	}}
}
