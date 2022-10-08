package pods

import (
	"context"
	"fmt"
	"kubeui/internal/app/pods/message"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newError(err error) tea.Cmd {
	return func() tea.Msg {
		return err
	}
}

func (m Model) listNamespaces() tea.Msg {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	namespaces, err := m.kubectl.CoreV1().Namespaces().List(ctx, v1.ListOptions{})

	if err != nil {
		return fmt.Errorf("failed to list namespaces: %v", err)
	}

	return message.NewInitialization(namespaces)
}

func (m Model) listPods() tea.Msg {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pods, err := m.kubectl.CoreV1().Pods(m.currentNamespace).List(ctx, v1.ListOptions{})

	if err != nil {
		return fmt.Errorf("failed to list namespaces: %v", err)
	}

	return message.NewListPods(pods)
}
