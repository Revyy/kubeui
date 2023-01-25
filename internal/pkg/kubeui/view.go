// Package kubui provides a framework that you can use to build view based applications requiring access to kubernetes.
// It is not meant to be a generic framework for building bubbletea based applications, instead if aims to solve specific problems
// faced when building the kubui program.
package kubeui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// View is intended to be a stateful component that completely
// takes over the ui and handles most inputs except for some global keypresses and system messages.
type View interface {
	Init(Context) tea.Cmd
	Update(Context, Msg) (Context, View, tea.Cmd)
	View(Context) string
	Destroy(Context) tea.Cmd
}
