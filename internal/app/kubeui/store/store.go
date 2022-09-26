package store

import (
	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type Store struct {
	Namespaces        []v1.Namespace
	SelectedNamespace v1.Namespace
	Logger            *zap.Logger
	WindowSize        tea.WindowSizeMsg
	Kubectl           *kubernetes.Clientset
}

func (s *Store) SetLogger(logger *zap.Logger) *Store {
	s.Logger = logger
	return s
}

func (s *Store) SetWindowSize(size tea.WindowSizeMsg) *Store {
	s.WindowSize = size
	return s
}

func (s *Store) SetNamespaces(namespaces []v1.Namespace) *Store {
	s.Namespaces = namespaces
	return s
}

func (s *Store) SetSelectedNamespace(namespace v1.Namespace) *Store {
	s.SelectedNamespace = namespace
	return s
}

func New(k *kubernetes.Clientset) *Store {
	return &Store{Kubectl: k}
}
