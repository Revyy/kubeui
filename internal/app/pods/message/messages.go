package message

import v1 "k8s.io/api/core/v1"

// Initialization is sent after fetching the initial list of available namespaces.
type Initialization struct {
	NamespaceList *v1.NamespaceList
}

// NewInitialization creates a new Initialization message.
func NewInitialization(namespaceList *v1.NamespaceList) Initialization {
	return Initialization{NamespaceList: namespaceList}
}

// ListPods is used as the result of fetching a list of pods in the current namespace.
type ListPods struct {
	PodList *v1.PodList
}

// NewListPods creates a new ListPods message.
func NewListPods(podList *v1.PodList) ListPods {
	return ListPods{PodList: podList}
}

// PodDeleted is sent after a pod has been deleted.
type PodDeleted struct {
	Name string
}

// NewPodDeleted creates a new PodDeleted message.
func NewPodDeleted(name string) PodDeleted {
	return PodDeleted{Name: name}
}
