package pods

import (
	"context"
	"fmt"
	"kubeui/internal/app/pods/message"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// listNamespaces fetches all namespaces for the current context.
func (m Model) listNamespaces() tea.Msg {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	namespaces, err := m.kubectl.CoreV1().Namespaces().List(ctx, v1.ListOptions{})

	if err != nil {
		return fmt.Errorf("failed to list namespaces: %v", err)
	}

	return message.NewInitialization(namespaces)
}

// listPods fetches all pods for the current context and namespace.
func (m Model) listPods() tea.Msg {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pods, err := m.kubectl.CoreV1().Pods(m.currentNamespace).List(ctx, v1.ListOptions{})

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

		pod, err := m.kubectl.CoreV1().Pods(m.currentNamespace).Get(ctx, id, v1.GetOptions{})

		if err != nil {
			return fmt.Errorf("failed to get pod: %v", err)
		}

		eventsCtx, eventsCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer eventsCancel()

		events, err := m.kubectl.CoreV1().Events(m.currentNamespace).List(eventsCtx, v1.ListOptions{FieldSelector: fmt.Sprintf("involvedObject.name=%s", pod.Name), TypeMeta: v1.TypeMeta{Kind: "Pod"}})

		if err != nil {
			return fmt.Errorf("failed to get pod events: %v", err)
		}

		return message.NewGetPod(pod, events.Items)
	}

}

// deletePod deletes a pod in the current context and namespace.
func (m Model) deletePod(name string) tea.Cmd {

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		err := m.kubectl.CoreV1().Pods(m.currentNamespace).Delete(ctx, name, v1.DeleteOptions{})

		if err != nil {
			return fmt.Errorf("failed to delete pod: %v", err)
		}

		return message.NewPodDeleted(name)
	}

}
