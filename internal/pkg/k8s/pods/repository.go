package pods

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// Repository defines the interface for the pods repository.
type Repository interface {
	Get(ctx context.Context, namespace, name string) (*v1.Pod, error)
	Delete(ctx context.Context, namespace, name string) error
	List(ctx context.Context, namespace string) (*v1.PodList, error)
	Events(ctx context.Context, namespace, name string) (*v1.EventList, error)
	TailLogs(ctx context.Context, pod *v1.Pod, options LogsOptions) (map[string]string, error)
}

// NewRepository creates a new Client.
func NewRepository(kubectl corev1.CoreV1Interface) Repository {
	return &ClientImpl{
		kubectl: kubectl,
	}
}

// ClientImpl is used to fetch pod related data from kubernetes.
type ClientImpl struct {
	kubectl corev1.CoreV1Interface
}

// Get fetches a single pod.
func (c *ClientImpl) Get(ctx context.Context, namespace, name string) (*v1.Pod, error) {
	return c.kubectl.Pods(namespace).Get(ctx, name, metav1.GetOptions{})
}

// Delete deletes a pod.
func (c *ClientImpl) Delete(ctx context.Context, namespace, name string) error {
	return c.kubectl.Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// List fetches a list of pods for a given namespace.
func (c *ClientImpl) List(ctx context.Context, namespace string) (*v1.PodList, error) {
	return c.kubectl.Pods(namespace).List(ctx, metav1.ListOptions{})
}

// Events fetches the current events for a pod.
func (c *ClientImpl) Events(ctx context.Context, namespace, name string) (*v1.EventList, error) {
	return c.kubectl.Events(namespace).List(ctx, metav1.ListOptions{FieldSelector: fmt.Sprintf("involvedObject.name=%s", name), TypeMeta: metav1.TypeMeta{Kind: "Pod"}})
}

// LogsOptions defines extra options to apply when fetching logs.
type LogsOptions struct {
	Count uint32
}

// TailLogs fetches the latest logs for a pod.
// By default the last 100 log lines will be fetched but this can be customized using the options parameter.
// Logs are returned as a mapping between the container name and the logs as a unified string separated by linebreaks '\n'.
func (c *ClientImpl) TailLogs(ctx context.Context, pod *v1.Pod, options LogsOptions) (map[string]string, error) {
	containerLogs := map[string]string{}

	errGroup := &errgroup.Group{}

	for i := range pod.Status.ContainerStatuses {

		container := pod.Status.ContainerStatuses[i]

		errGroup.Go(func() error {

			logsCtx, logsCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer logsCancel()

			tailLines := int64(options.Count)

			if tailLines == 0 {
				tailLines = 100
			}

			logsRequest := c.kubectl.Pods(pod.Namespace).GetLogs(pod.GetName(), &v1.PodLogOptions{Container: container.Name, TailLines: &tailLines})

			if logsRequest == nil {
				return fmt.Errorf("failed to issue request to fetch container logs for %s", container.Name)
			}

			podLogs, err := logsRequest.Stream(logsCtx)

			if err != nil {
				return err
			}
			defer podLogs.Close()

			buf := new(bytes.Buffer)
			_, err = io.Copy(buf, podLogs)
			if err != nil {
				return err
			}

			containerLogs[container.Name] = buf.String()

			return nil
		})
	}

	if err := errGroup.Wait(); err != nil {
		return containerLogs, err
	}

	return containerLogs, nil
}
