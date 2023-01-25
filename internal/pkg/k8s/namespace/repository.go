package namespace

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// Repository defines the interface for the pods client.
type Repository interface {
	List(ctx context.Context) (*v1.NamespaceList, error)
}

// NewRepository creates a new Client.
func NewRepository(kubectl corev1.CoreV1Interface) Repository {
	return &RepositoryImpl{
		kubectl: kubectl,
	}
}

// RepositoryImpl is used to fetch pod related data from kubernetes.
type RepositoryImpl struct {
	kubectl corev1.CoreV1Interface
}

// List lists all namespaces.
func (c *RepositoryImpl) List(ctx context.Context) (*v1.NamespaceList, error) {
	return c.kubectl.Namespaces().List(ctx, metav1.ListOptions{})
}
