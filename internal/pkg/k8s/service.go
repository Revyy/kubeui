package k8s

import (
	"context"
	"fmt"
	"kubeui/internal/pkg/k8s/namespace"
	"kubeui/internal/pkg/k8s/pods"
	"time"

	v1 "k8s.io/api/core/v1"
)

// Service defines the interface to fetch data from kubernetes.
// It might be split up into several more specialized services in the future.
type Service interface {
	// Lists namespaces in the cluster.
	ListNamespaces() (*v1.NamespaceList, error)
	// Lists pods in the specified namespace.
	ListPods(namespace string) (*v1.PodList, error)
	// Fetches information about a single pod, including events and logs.
	GetPod(namespace, id string) (*Pod, error)
	// Delete the pod with the specified name in the specified namespace.
	// Returns the name of the deleted pod.
	DeletePod(namespace, name string) (string, error)
}

// NewK8sService creates a new K8sClient.
func NewK8sService(podsRepository pods.Repository, namespaceRepository namespace.Repository) Service {
	return &K8sServiceImpl{
		PodsRepository:      podsRepository,
		NamespaceRepository: namespaceRepository,
	}
}

// K8sServiceImpl is used to fetch data and issue commands to a kubernetes cluster.
type K8sServiceImpl struct {
	PodsRepository      pods.Repository
	NamespaceRepository namespace.Repository
}

// ListNamespaces fetches all namespaces for the current context.
func (c *K8sServiceImpl) ListNamespaces() (*v1.NamespaceList, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	namespaces, err := c.NamespaceRepository.List(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %v", err)
	}

	return namespaces, nil

}

// ListPods fetches all pods for the current context and namespace.
func (c *K8sServiceImpl) ListPods(namespace string) (*v1.PodList, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pods, err := c.PodsRepository.List(ctx, namespace)

	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %v", err)
	}

	return pods, nil

}

// GetPod fetches a pod in the current context and namespace.
func (c *K8sServiceImpl) GetPod(namespace, name string) (*Pod, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pod, err := c.PodsRepository.Get(ctx, namespace, name)

	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %v", err)
	}

	eventsCtx, eventsCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer eventsCancel()

	events, err := c.PodsRepository.Events(eventsCtx, namespace, name)

	if err != nil {
		return nil, fmt.Errorf("failed to get pod events: %v", err)
	}

	logs := map[string]string{}

	logsCtx, logsCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer logsCancel()

	if len(pod.Status.ContainerStatuses) > 0 {
		logs, err = c.PodsRepository.TailLogs(logsCtx, pod, pods.LogsOptions{})
	}

	if err != nil {
		return nil, err
	}

	return &Pod{
		Pod:    *pod,
		Events: events.Items,
		Logs:   logs,
	}, nil

}

// DeletePod deletes a pod in the current context and namespace.
func (c *K8sServiceImpl) DeletePod(namespace, name string) (string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := c.PodsRepository.Delete(ctx, namespace, name)

	if err != nil {
		return "", fmt.Errorf("failed to delete pod: %v", err)
	}

	return name, nil
}
