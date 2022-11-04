package k8scommand

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ListNamespaces fetches all namespaces for the current context.
func ListNamespaces(kubectl *kubernetes.Clientset) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		namespaces, err := kubectl.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})

		if err != nil {
			return fmt.Errorf("failed to list namespaces: %v", err)
		}

		return NewListNamespacesMsg(namespaces)
	}
}

// ListPods fetches all pods for the current context and namespace.
func ListPods(kubectl *kubernetes.Clientset, namespace string) tea.Cmd {

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		pods, err := kubectl.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})

		if err != nil {
			return fmt.Errorf("failed to list namespaces: %v", err)
		}

		return NewListPodsMsg(pods)
	}
}

// GetPod fetches a pod in the current context and namespace.
func GetPod(kubectl *kubernetes.Clientset, namespace, id string) tea.Cmd {

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		pod, err := kubectl.CoreV1().Pods(namespace).Get(ctx, id, metav1.GetOptions{})

		if err != nil {
			return fmt.Errorf("failed to get pod: %v", err)
		}

		eventsCtx, eventsCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer eventsCancel()

		events, err := kubectl.CoreV1().Events(namespace).List(eventsCtx, metav1.ListOptions{FieldSelector: fmt.Sprintf("involvedObject.name=%s", pod.Name), TypeMeta: metav1.TypeMeta{Kind: "Pod"}})

		if err != nil {
			return fmt.Errorf("failed to get pod events: %v", err)
		}

		logs := ""

		if len(pod.Status.ContainerStatuses) > 0 {
			logs, err = getLogs(kubectl, namespace, pod.Name, pod.Status.ContainerStatuses[0].Name)
		}

		if err != nil {
			return err
		}

		return NewGetPodMsg(pod, events.Items, logs)
	}

}

func getLogs(kubectl *kubernetes.Clientset, namespace, podName, containerName string) (string, error) {

	logsCtx, logsCancel := context.WithTimeout(context.Background(), 5*time.Second)
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

// DeletePod deletes a pod in the current context and namespace.
func DeletePod(kubectl *kubernetes.Clientset, namespace, name string) tea.Cmd {

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := kubectl.CoreV1().Pods(name).Delete(ctx, name, metav1.DeleteOptions{})

		if err != nil {
			return fmt.Errorf("failed to delete pod: %v", err)
		}

		return NewPodDeletedMsg(name)
	}

}
