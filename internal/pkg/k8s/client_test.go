package k8s_test

import (
	"fmt"
	"kubeui/internal/pkg/k8s"
	"testing"

	"github.com/life4/genesis/maps"
	"github.com/stretchr/testify/assert"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

func TestContextClientImpl_CurrentApiContext(t *testing.T) {

	createConfig := func(currentContext string, contexts map[string]*api.Context) api.Config {
		config := api.NewConfig()
		config.CurrentContext = currentContext
		config.Contexts = contexts
		return *config
	}

	tests := []struct {
		name         string
		config       api.Config
		configAccess clientcmd.ConfigAccess
		want         *api.Context
		exist        bool
	}{
		{
			"Empty current context",
			createConfig("", map[string]*api.Context{"test": api.NewContext()}),
			nil,
			nil,
			false,
		},
		{
			"Non existing current context",
			createConfig("not-here", map[string]*api.Context{"test": api.NewContext()}),
			nil,
			nil,
			false,
		},
		{
			"Successfully found current context",
			createConfig("test", map[string]*api.Context{"test": api.NewContext()}),
			nil,
			api.NewContext(),
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := k8s.NewContextClientImpl(tt.configAccess, tt.config, nil)

			got, exist := c.CurrentApiContext()
			assert.EqualValues(t, got, tt.want)
			assert.Equal(t, exist, tt.exist)

		})
	}
}

func TestContextClientImpl_CurrentContext(t *testing.T) {

	createConfig := func(currentContext string) api.Config {
		config := api.NewConfig()
		config.CurrentContext = currentContext
		return *config
	}

	tests := []struct {
		name         string
		config       api.Config
		configAccess clientcmd.ConfigAccess
		want         string
	}{
		{"No current context set", createConfig(""), nil, ""},
		{"Default context", api.Config{}, nil, ""},
		{"Specific context set", createConfig("test-context"), nil, "test-context"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := k8s.NewContextClientImpl(tt.configAccess, tt.config, nil)
			got := c.CurrentContext()
			assert.EqualValues(t, got, tt.want)
		})
	}
}

func TestContextClientImpl_Contexts(t *testing.T) {

	createConfig := func(contexts map[string]*api.Context) api.Config {
		config := api.NewConfig()
		config.Contexts = contexts
		return *config
	}

	tests := []struct {
		name         string
		config       api.Config
		configAccess clientcmd.ConfigAccess
		want         []string
	}{
		{"Default config should return empty slice", api.Config{}, nil, []string{}},
		{"No configs should return empty slice", createConfig(map[string]*api.Context{}), nil, []string{}},
		{"Should return actual context keys", createConfig(map[string]*api.Context{
			"test":  api.NewContext(),
			"test2": api.NewContext(),
		}), nil, []string{"test", "test2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := k8s.NewContextClientImpl(tt.configAccess, tt.config, nil)
			got := c.Contexts()
			assert.Subset(t, got, tt.want)
			assert.Equal(t, len(tt.want), len(got))
		})
	}
}

func TestContextClientImpl_SwitchContext(t *testing.T) {

	nilModifyFunc := func(configAccess clientcmd.ConfigAccess, newConfig api.Config, relativizePaths bool) error {
		return nil
	}

	errModifyFunc := func(configAccess clientcmd.ConfigAccess, newConfig api.Config, relativizePaths bool) error {
		return fmt.Errorf("some error")
	}

	createConfig := func(currentContext string, contexts map[string]*api.Context) api.Config {
		config := api.NewConfig()
		config.CurrentContext = currentContext
		config.Contexts = contexts
		return *config
	}

	type fields struct {
		modifyConfig   k8s.ModifyConfigFunc
		configAccess   clientcmd.ConfigAccess
		currentContext string
		contexts       map[string]*api.Context
	}

	type args struct {
		ctx       string
		namespace string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"Non existing context",
			fields{nilModifyFunc, nil, "test", map[string]*api.Context{"test": api.NewContext()}},
			args{"not-there", "default"},
			true,
		},
		// This should be valid input except that ModifyFunc returns an err
		{
			"ModifyFunc returns err",
			fields{errModifyFunc, nil, "test", map[string]*api.Context{
				"test":  api.NewContext(),
				"test2": api.NewContext(),
			}},
			args{"test2", "default"},
			true,
		},
		{
			"Successful switch",
			fields{nilModifyFunc, nil, "test", map[string]*api.Context{
				"test":  api.NewContext(),
				"test2": api.NewContext(),
			}},
			args{"test2", "default"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := createConfig(tt.fields.currentContext, tt.fields.contexts)
			c := k8s.NewContextClientImpl(tt.fields.configAccess, config, tt.fields.modifyConfig)
			err := c.SwitchContext(tt.args.ctx, tt.args.namespace)

			if tt.wantErr {
				assert.Error(t, err)

				// If we get an error, then the current context should not have changed from our default.
				assert.Equal(t, tt.fields.currentContext, c.CurrentContext())

				// If we get an error, then the namespace of the current context should not have changed.
				current, _ := c.CurrentApiContext()
				assert.Equal(t, "", current.Namespace)
			} else {
				assert.Nil(t, err)
				// Check that the current context matches what we supplied.
				assert.Equal(t, tt.args.ctx, c.CurrentContext())

				// If we supplied a namespace we expect the namespace of the current context to match.
				if tt.args.namespace != "" {
					current, _ := c.CurrentApiContext()
					assert.Equal(t, tt.args.namespace, current.Namespace)
				}
			}
		})
	}
}

func TestContextClientImpl_DeleteContext(t *testing.T) {

	nilModifyFunc := func(configAccess clientcmd.ConfigAccess, newConfig api.Config, relativizePaths bool) error {
		return nil
	}

	errModifyFunc := func(configAccess clientcmd.ConfigAccess, newConfig api.Config, relativizePaths bool) error {
		return fmt.Errorf("some error")
	}

	createConfig := func(currentContext string, contexts map[string]*api.Context) api.Config {
		config := api.NewConfig()
		config.CurrentContext = currentContext
		config.Contexts = contexts
		return *config
	}

	type fields struct {
		modifyConfig   k8s.ModifyConfigFunc
		configAccess   clientcmd.ConfigAccess
		currentContext string
		contexts       map[string]*api.Context
	}

	tests := []struct {
		name    string
		fields  fields
		ctx     string
		wantErr bool
	}{
		{
			"Non existing context",
			fields{nilModifyFunc, nil, "test", map[string]*api.Context{"test": api.NewContext()}},
			"not-there",
			true,
		},
		// This should be valid input except that ModifyFunc returns an err
		{
			"ModifyFunc returns err",
			fields{errModifyFunc, nil, "test", map[string]*api.Context{
				"test":  api.NewContext(),
				"test2": api.NewContext(),
			}},
			"test2",
			true,
		},
		{
			"Successful deletion",
			fields{nilModifyFunc, nil, "test", map[string]*api.Context{
				"test":  api.NewContext(),
				"test2": api.NewContext(),
			}},
			"test2",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := createConfig(tt.fields.currentContext, tt.fields.contexts)
			originalContextKeys := maps.Keys(config.Contexts)

			c := k8s.NewContextClientImpl(tt.fields.configAccess, config, tt.fields.modifyConfig)
			err := c.DeleteContext(tt.ctx)

			if tt.wantErr {
				assert.Error(t, err)
				// If we get an error we expect the contexts to match our default ones.
				// Meaning no context was deleted.
				assert.Subset(t, config.Contexts, originalContextKeys)
				assert.Equal(t, len(config.Contexts), len(originalContextKeys))
			} else {
				assert.Nil(t, err)
				assert.NotContains(t, config.Contexts, tt.ctx)
			}
		})
	}

	t.Run("Deleting the current context", func(t *testing.T) {
		config := createConfig("test", map[string]*api.Context{
			"test": api.NewContext(),
		})

		c := k8s.NewContextClientImpl(nil, config, nilModifyFunc)
		err := c.DeleteContext("test")

		assert.Nil(t, err)
		// The current context should be reset to the empty string.
		assert.Equal(t, "", c.CurrentContext())
		assert.NotContains(t, c.Contexts(), "")
	})
}

func TestContextClientImpl_DeleteUser(t *testing.T) {

	nilModifyFunc := func(configAccess clientcmd.ConfigAccess, newConfig api.Config, relativizePaths bool) error {
		return nil
	}

	errModifyFunc := func(configAccess clientcmd.ConfigAccess, newConfig api.Config, relativizePaths bool) error {
		return fmt.Errorf("some error")
	}

	createConfig := func(authInfos map[string]*api.AuthInfo) api.Config {
		config := api.NewConfig()
		config.AuthInfos = authInfos
		return *config
	}

	type fields struct {
		modifyConfig k8s.ModifyConfigFunc
		configAccess clientcmd.ConfigAccess
		authInfos    map[string]*api.AuthInfo
	}

	tests := []struct {
		name    string
		fields  fields
		user    string
		wantErr bool
	}{
		{
			"Non existing context",
			fields{nilModifyFunc, nil, map[string]*api.AuthInfo{"test": api.NewAuthInfo()}},
			"not-there",
			true,
		},
		// This should be valid input except that ModifyFunc returns an err
		{
			"ModifyFunc returns err",
			fields{errModifyFunc, nil, map[string]*api.AuthInfo{
				"test":  api.NewAuthInfo(),
				"test2": api.NewAuthInfo(),
			}},
			"test2",
			true,
		},
		{
			"Successful deletion",
			fields{nilModifyFunc, nil, map[string]*api.AuthInfo{
				"test":  api.NewAuthInfo(),
				"test2": api.NewAuthInfo(),
			}},
			"test2",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := createConfig(tt.fields.authInfos)
			originalAuthInfoKeys := maps.Keys(config.AuthInfos)

			c := k8s.NewContextClientImpl(tt.fields.configAccess, config, tt.fields.modifyConfig)
			err := c.DeleteUser(tt.user)

			if tt.wantErr {
				assert.Error(t, err)
				// If we get an error we expect the AuthInfos to match our default ones.
				// Meaning no user was deleted.
				assert.Subset(t, config.AuthInfos, originalAuthInfoKeys)
				assert.Equal(t, len(config.AuthInfos), len(originalAuthInfoKeys))
			} else {
				assert.Nil(t, err)
				assert.NotContains(t, config.AuthInfos, tt.user)
			}
		})
	}
}

func TestContextClientImpl_DeleteClusterEntry(t *testing.T) {

	nilModifyFunc := func(configAccess clientcmd.ConfigAccess, newConfig api.Config, relativizePaths bool) error {
		return nil
	}

	errModifyFunc := func(configAccess clientcmd.ConfigAccess, newConfig api.Config, relativizePaths bool) error {
		return fmt.Errorf("some error")
	}

	createConfig := func(clusters map[string]*api.Cluster) api.Config {
		config := api.NewConfig()
		config.Clusters = clusters
		return *config
	}

	type fields struct {
		modifyConfig k8s.ModifyConfigFunc
		configAccess clientcmd.ConfigAccess
		clusters     map[string]*api.Cluster
	}

	tests := []struct {
		name    string
		fields  fields
		cluster string
		wantErr bool
	}{
		{
			"Non existing context",
			fields{nilModifyFunc, nil, map[string]*api.Cluster{"test": api.NewCluster()}},
			"not-there",
			true,
		},
		// This should be valid input except that ModifyFunc returns an err
		{
			"ModifyFunc returns err",
			fields{errModifyFunc, nil, map[string]*api.Cluster{
				"test":  api.NewCluster(),
				"test2": api.NewCluster(),
			}},
			"test2",
			true,
		},
		{
			"Successful deletion",
			fields{nilModifyFunc, nil, map[string]*api.Cluster{
				"test":  api.NewCluster(),
				"test2": api.NewCluster(),
			}},
			"test2",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := createConfig(tt.fields.clusters)
			originalClusterKeys := maps.Keys(config.Clusters)

			c := k8s.NewContextClientImpl(tt.fields.configAccess, config, tt.fields.modifyConfig)
			err := c.DeleteClusterEntry(tt.cluster)

			if tt.wantErr {
				assert.Error(t, err)
				// If we get an error we expect the AuthInfos to match our default ones.
				// Meaning no user was deleted.
				assert.Subset(t, config.Clusters, originalClusterKeys)
				assert.Equal(t, len(config.Clusters), len(originalClusterKeys))
			} else {
				assert.Nil(t, err)
				assert.NotContains(t, config.Clusters, tt.cluster)
			}
		})
	}
}
