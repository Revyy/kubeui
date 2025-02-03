package kubeui_test

import (
	"fmt"
	"testing"

	"kubeui/internal/pkg/kubeui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestExit(t *testing.T) {
	cmd := kubeui.Exit()
	assert.Equal(t, tea.Quit(), cmd())
}

func TestError(t *testing.T) {
	err := fmt.Errorf("error")
	cmd := kubeui.Error(err)

	assert.Equal(t, err, cmd())
}

func TestExitView(t *testing.T) {
	cmd := kubeui.ExitView()
	assert.Equal(t, kubeui.ExitViewMsg{}, cmd())
}

func TestPushView(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		initialize bool
		want       kubeui.PushViewMsg
	}{
		{"initialize false", "some-view", false, kubeui.PushViewMsg{Id: "some-view", Initialize: false}},
		{"initialize true", "some-view2", true, kubeui.PushViewMsg{Id: "some-view2", Initialize: true}},
		{"empty id, initialize false", "", false, kubeui.PushViewMsg{Id: "", Initialize: false}},
		{"empty id, initialize true", "", true, kubeui.PushViewMsg{Id: "", Initialize: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := kubeui.PushView(tt.id, tt.initialize)
			assert.Equal(t, tt.want, cmd())
		})
	}
}

func TestPopView(t *testing.T) {
	tests := []struct {
		name       string
		initialize bool
		want       kubeui.PopViewMsg
	}{
		{"initialize false", false, kubeui.PopViewMsg{Initialize: false}},
		{"initialize true", true, kubeui.PopViewMsg{Initialize: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := kubeui.PopView(tt.initialize)
			assert.Equal(t, tt.want, cmd())
		})
	}
}
