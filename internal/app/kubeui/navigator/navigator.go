package navigator

import (
	"fmt"
	"kubeui/internal/app/kubeui/page"
	"kubeui/internal/app/kubeui/store"
)

// Defines a route in the application.
// Given the current state(store) and a set of parameters it should return a page.
type NavFunc func(*store.Store, map[string]string) page.Page

type RouteMsg struct {
	Name       string
	Store      *store.Store
	Parameters map[string]string
}

// Navigator is a basic pagenavigator.
type Navigator struct {
	currentPage page.Page
	pages       map[string]NavFunc
}

func New(initialPage page.Page) *Navigator {
	return &Navigator{
		currentPage: initialPage,
		pages:       map[string]NavFunc{},
	}
}

func (n *Navigator) Add(name string, navFunc NavFunc) {
	n.pages[name] = navFunc
}

func (n *Navigator) Current() page.Page {
	return n.currentPage
}

func (n *Navigator) Navigate(to string, store *store.Store, parameters map[string]string) (page.Page, error) {
	navFunc, ok := n.pages[to]

	if !ok {
		return nil, fmt.Errorf("page %s is not registered", to)
	}

	n.currentPage = navFunc(store, parameters)
	return n.currentPage, nil
}
