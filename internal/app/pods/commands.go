package pods

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"kubeui/internal/app/pods/message"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// listNamespaces fetches all namespaces for the current context.
func (m Model) listNamespaces() tea.Msg {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	namespaces, err := m.kubectl.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})

	if err != nil {
		return fmt.Errorf("failed to list namespaces: %v", err)
	}

	return message.NewInitialization(namespaces)
}

// listPods fetches all pods for the current context and namespace.
func (m Model) listPods() tea.Msg {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pods, err := m.kubectl.CoreV1().Pods(m.currentNamespace).List(ctx, metav1.ListOptions{})

	if err != nil {
		return fmt.Errorf("failed to list namespaces: %v", err)
	}

	return message.NewListPods(pods)
}

// getPod fetches a pod in the current context and namespace.
func (m Model) getPod(id string) tea.Cmd {

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		pod, err := m.kubectl.CoreV1().Pods(m.currentNamespace).Get(ctx, id, metav1.GetOptions{})

		if err != nil {
			return fmt.Errorf("failed to get pod: %v", err)
		}

		eventsCtx, eventsCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer eventsCancel()

		events, err := m.kubectl.CoreV1().Events(m.currentNamespace).List(eventsCtx, metav1.ListOptions{FieldSelector: fmt.Sprintf("involvedObject.name=%s", pod.Name), TypeMeta: metav1.TypeMeta{Kind: "Pod"}})

		if err != nil {
			return fmt.Errorf("failed to get pod events: %v", err)
		}

		logs := ""

		if len(pod.Status.ContainerStatuses) > 0 {
			logs, err = getLogs(m.kubectl, m.currentNamespace, pod.Name, pod.Status.ContainerStatuses[0].Name)
		}

		if err != nil {
			return err
		}

		return message.NewGetPod(pod, events.Items, logs)
	}

}

func getLogs(kubectl *kubernetes.Clientset, namespace, podName, containerName string) (string, error) {

	logsCtx, logsCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer logsCancel()

	tailLines := int64(100)
	logsRequest := kubectl.CoreV1().Pods(namespace).GetLogs(podName, &v1.PodLogOptions{Container: containerName, TailLines: &tailLines})

	podLogs, err := logsRequest.Stream(logsCtx)

	if err != nil {
		return "", err
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// deletePod deletes a pod in the current context and namespace.
func (m Model) deletePod(name string) tea.Cmd {

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		err := m.kubectl.CoreV1().Pods(m.currentNamespace).Delete(ctx, name, metav1.DeleteOptions{})

		if err != nil {
			return fmt.Errorf("failed to delete pod: %v", err)
		}

		return message.NewPodDeleted(name)
	}

}
