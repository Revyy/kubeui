package namespace

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// Client defines the interface for the pods client.
type Client interface {
	List(ctx context.Context) (*v1.NamespaceList, error)
}

// NewClient creates a new Client.
func NewClient(kubectl corev1.CoreV1Interface) Client {
	return &ClientImpl{
		kubectl: kubectl,
	}
}

// ClientImpl is used to fetch pod related data from kubernetes.
type ClientImpl struct {
	kubectl corev1.CoreV1Interface
}

// List lists all namespaces.
func (c *ClientImpl) List(ctx context.Context) (*v1.NamespaceList, error) {
	return c.kubectl.Namespaces().List(ctx, metav1.ListOptions{})
}
