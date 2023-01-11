package k8scontext

import (
	"fmt"

	"github.com/life4/genesis/maps"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Client defines the interface for working with kubernetes contexts from a view.
type Client interface {
	// Returns a list of available contexts.
	Contexts() []string
	// Returns the api.Context for the currently selected context if it exists.
	// If no api.Context exists for the current context then the bool should be set to false.
	CurrentApiContext() (*api.Context, bool)
	// Returns the currently selected context.
	CurrentContext() string
	// Switch to the specified context and optionally set the default namespace.
	SwitchContext(ctx, namespace string) (err error)
	// Delete the specified context.
	DeleteContext(ctx string) (err error)
	// Delete the specified user entry.
	DeleteUser(user string) (err error)
	// Delete the specified cluster entry.
	DeleteClusterEntry(cluster string) (err error)
}

// ModifyConfigFunc is used to modify the underlying configuration in the file-system.
// This exists to enable testing the ContextClientImpl without manipulating actual files in the filesystem.
type ModifyConfigFunc func(configAccess clientcmd.ConfigAccess, newConfig api.Config, relativizePaths bool) error

// ClientImpl is used to manipulate the current context.
type ClientImpl struct {
	modifyConfig ModifyConfigFunc
	configAccess clientcmd.ConfigAccess
	config       api.Config
}

// NewClientImpl creates a new Client.
//
// If modifyConfig is nil it will default to clientcmd.ModifyConfig.
func NewClientImpl(configAccess clientcmd.ConfigAccess, config api.Config, modifyConfig ModifyConfigFunc) *ClientImpl {
	impl := &ClientImpl{
		configAccess: configAccess,
		config:       config,
		modifyConfig: clientcmd.ModifyConfig,
	}

	if modifyConfig != nil {
		impl.modifyConfig = modifyConfig
	}

	return impl

}

// CurrentApiContext returns the currently active api context.
func (c *ClientImpl) CurrentApiContext() (*api.Context, bool) {
	ctx, ok := c.config.Contexts[c.config.CurrentContext]
	return ctx, ok
}

// CurrentContext returns the currently active context.
func (c *ClientImpl) CurrentContext() string {
	return c.config.CurrentContext
}

// Contexts returns a list of available contexts.
func (c *ClientImpl) Contexts() []string {
	return maps.Keys(c.config.Contexts)
}

// SwitchContext changes the active context in a kubeconfig.
func (c *ClientImpl) SwitchContext(ctx, namespace string) (err error) {

	kubeCtx, ok := c.config.Contexts[ctx]

	if !ok {
		return fmt.Errorf("context %s doesn't exists", ctx)
	}

	oldContext := c.config.CurrentContext
	c.config.CurrentContext = ctx

	err = c.modifyConfig(c.configAccess, c.config, true)
	if err != nil {
		// Restore context if we fail to modify the config.
		c.config.CurrentContext = oldContext
		return fmt.Errorf("error ModifyConfig: %v", err)
	}

	if namespace != "" {
		kubeCtx.Namespace = namespace
	}

	return nil
}

// DeleteContext deletes a context from a kubeconfig file.
func (c *ClientImpl) DeleteContext(ctx string) (err error) {

	_, ok := c.config.Contexts[ctx]
	if !ok {
		return fmt.Errorf("context %s doesn't exists", ctx)
	}

	if err := c.modifyConfig(c.configAccess, c.config, true); err != nil {
		return err
	}

	delete(c.config.Contexts, ctx)

	if ctx == c.config.CurrentContext {
		c.config.CurrentContext = ""
	}

	return nil
}

// DeleteUser deletes a user entry from a kubeconfig file.
func (c *ClientImpl) DeleteUser(user string) (err error) {

	_, ok := c.config.AuthInfos[user]
	if !ok {
		return fmt.Errorf("user %s doesn't exists", user)
	}

	if err := c.modifyConfig(c.configAccess, c.config, true); err != nil {
		return err
	}

	delete(c.config.AuthInfos, user)

	return nil
}

// DeleteClusterEntry deletes a cluster entry from a kubeconfig file.
func (c *ClientImpl) DeleteClusterEntry(cluster string) (err error) {

	_, ok := c.config.Clusters[cluster]
	if !ok {
		return fmt.Errorf("cluster %s doesn't exists", cluster)
	}

	if err := c.modifyConfig(c.configAccess, c.config, true); err != nil {
		return err
	}

	delete(c.config.Clusters, cluster)

	return nil
}
