package pods

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (m Model) getNamespaces() tea.Msg {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	namespaces, err := m.kubectl.CoreV1().Namespaces().List(ctx, v1.ListOptions{})

	if err != nil {
		return fmt.Errorf("failed to list namespaces: %v", err)
	}

	return namespaces
}
