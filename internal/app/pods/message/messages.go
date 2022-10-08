package message

import v1 "k8s.io/api/core/v1"

type Initialization struct {
	NamespaceList *v1.NamespaceList
}

func NewInitialization(namespaceList *v1.NamespaceList) Initialization {
	return Initialization{NamespaceList: namespaceList}
}

type ListPods struct {
	PodList *v1.PodList
}

func NewListPods(podList *v1.PodList) ListPods {
	return ListPods{PodList: podList}
}

type PodDeleted struct {
	Name string
}

func NewPodDeleted(name string) PodDeleted {
	return PodDeleted{Name: name}
}
