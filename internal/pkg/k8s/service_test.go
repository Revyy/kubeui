package k8s_test

import (
	"context"
	"fmt"
	"kubeui/internal/pkg/k8s"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

type mockNamespaceRepository struct {
	namespaceList *v1.NamespaceList
	err           error
}

func (c *mockNamespaceRepository) List(ctx context.Context) (*v1.NamespaceList, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.namespaceList, nil
}

type listNamespacesTest struct {
	name       string
	repository *mockNamespaceRepository
	wantErr    bool
	expected   *v1.NamespaceList
}

var listNamespacesTests = []listNamespacesTest{
	{"Namespace client returns error", &mockNamespaceRepository{nil, fmt.Errorf("error")}, true, nil},
	{"Namespace client returns empty list", &mockNamespaceRepository{&v1.NamespaceList{}, nil}, false, &v1.NamespaceList{}},
	{"Namespace client returns non empty list", &mockNamespaceRepository{
		&v1.NamespaceList{
			Items: []v1.Namespace{
				{Spec: v1.NamespaceSpec{
					Finalizers: []v1.FinalizerName{
						"test",
					},
				}},
			},
		}, nil}, false,
		&v1.NamespaceList{
			Items: []v1.Namespace{
				{Spec: v1.NamespaceSpec{
					Finalizers: []v1.FinalizerName{
						"test",
					},
				}},
			},
		},
	},
}

func TestListNamespaces(t *testing.T) {
	for _, test := range listNamespacesTests {
		service := k8s.NewK8sService(nil, test.repository)
		got, err := service.ListNamespaces()

		if test.wantErr {
			assert.Error(t, err)
			assert.Nil(t, got)
		}

		if !test.wantErr {
			assert.Nil(t, err)
			assert.Equal(t, test.expected, got)
		}

	}
}
