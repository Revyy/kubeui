package kubeui

import (
	"kubeui/internal/pkg/k8s"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd/api"
)

// K8sClient defines the interface to fetch data from kubernetes.
// It might be split up into several more specialized clients in the future.
type K8sClient interface {
	// Lists namespaces in the cluster.
	ListNamespaces() (*v1.NamespaceList, error)
	// Lists pods in the specified namespace.
	ListPods(namespace string) (*v1.PodList, error)
	// Fetches information about a single pod, including events and logs.
	GetPod(namespace, id string) (*k8s.Pod, error)
	// Delete the pod with the specified name in the specified namespace.
	// Returns the name of the deleted pod.
	DeletePod(namespace, name string) (string, error)
}

// ContextClient defines the interface for working with kubernetes contexts from a view.
type ContextClient interface {
	// Returns a list of available contexts.
	Contexts() []string
	// Returns the api.Context for the currently selected context if it exists.
	// If no api.Context exists for the current context then the bool should be set to false.
	CurrentApiContext() (*api.Context, bool)
	// Returns the currently selected context.
	CurrentContext() string
	// Switch to the specified context and optionally set the default namespace.
	SwitchContext(ctx, namespace string) (err error)
	// Delete the specified context.
	DeleteContext(ctx string) (err error)
	// Delete the specified user entry.
	DeleteUser(user string) (err error)
	// Delete the specified cluster entry.
	DeleteClusterEntry(cluster string) (err error)
}

// Context contains the context of the kubeui application.
type Context struct {
	// Used to read and manipulate the kubeconfig file and related contexts.
	ContextClient ContextClient

	// Used to issue commands and fetch data from kubernetes.
	K8sClient K8sClient

	// Currently selected namespace
	Namespace string

	// Name of currently selected pod.
	SelectedPod string
}
