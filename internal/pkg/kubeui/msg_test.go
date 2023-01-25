package kubeui_test

import (
	"fmt"
	"kubeui/internal/pkg/kubeui"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestMsg_IsError(t *testing.T) {

	tests := []struct {
		name    string
		teaMsg  tea.Msg
		wantErr bool
	}{
		{"Should return error", fmt.Errorf("some error"), true},
		{"Should not return error", "some string message", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := kubeui.Msg{
				TeaMsg: tt.teaMsg,
			}

			got, ok := m.IsError()

			if tt.wantErr {
				assert.Error(t, got)
				assert.True(t, ok)
			} else {
				assert.Nil(t, got)
				assert.False(t, ok)
			}

		})
	}
}

func TestMsg_IsWindowResize(t *testing.T) {
	tests := []struct {
		name   string
		teaMsg tea.Msg
		want   bool
	}{
		{"Should not return true", 10, false},
		{"Should return true", tea.WindowSizeMsg{Width: 10, Height: 10}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := kubeui.Msg{
				TeaMsg: tt.teaMsg,
			}

			got := m.IsWindowResize()

			assert.Equal(t, tt.want, got)

		})
	}
}

func TestMsg_IsKeyMsg(t *testing.T) {
	tests := []struct {
		name   string
		teaMsg tea.Msg
		want   bool
	}{
		{"Should not return true", 10, false},
		{"Should return true", tea.KeyMsg{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := kubeui.Msg{
				TeaMsg: tt.teaMsg,
			}

			got := m.IsKeyMsg()

			assert.Equal(t, tt.want, got)

		})
	}
}

func TestMsg_MatchesKeyBindings(t *testing.T) {

	tests := []struct {
		name     string
		teaMsg   tea.Msg
		bindings []key.Binding
		want     bool
	}{
		{"Should return false is message is not a keyMsg", 10, []key.Binding{key.NewBinding(key.WithKeys("enter"))}, false},
		{"Should return false if keys do not match", tea.KeyMsg{Type: tea.KeyCtrlA}, []key.Binding{key.NewBinding(key.WithKeys(tea.KeyEnter.String()))}, false},
		{"Should return true if keys match", tea.KeyMsg{Type: tea.KeyEnter}, []key.Binding{key.NewBinding(key.WithKeys(tea.KeyEnter.String()))}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := kubeui.Msg{
				TeaMsg: tt.teaMsg,
			}

			got := m.MatchesKeyBindings(tt.bindings...)

			assert.Equal(t, tt.want, got)

		})
	}
}

func TestMsg_GetWindowResizeMsg(t *testing.T) {
	tests := []struct {
		name       string
		teaMsg     tea.Msg
		want       tea.WindowSizeMsg
		shouldFind bool
	}{
		{"Should return default value and false if msg is not of correct type", 10, tea.WindowSizeMsg{}, false},
		{"Should return correct value and true", tea.WindowSizeMsg{Width: 10, Height: 10}, tea.WindowSizeMsg{Width: 10, Height: 10}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := kubeui.Msg{
				TeaMsg: tt.teaMsg,
			}

			got, ok := m.GetWindowResizeMsg()

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.shouldFind, ok)

		})
	}
}
