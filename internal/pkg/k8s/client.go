package k8s

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/life4/genesis/maps"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// NewKClientSet creates a kubernetes ClientSet that can be used to issue kubernetes commands.
func NewKClientSet(kubeconfig string, access clientcmd.ConfigAccess) (*kubernetes.Clientset, error) {

	config, err := clientcmd.BuildConfigFromKubeconfigGetter("", access.GetStartingConfig)
	if err != nil {
		return nil, err
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

// NewClientConfig creates a ClientConfig object representing the kubeconfig of the user.
func NewClientConfig(context, kubeconfigPath string) clientcmd.ClientConfig {

	clientConfigLoadRules := clientcmd.NewDefaultPathOptions().LoadingRules

	if kubeconfigPath != "" {
		clientConfigLoadRules = &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath}
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientConfigLoadRules,
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		})
}

// Client defines the interface to fetch data from kubernetes.
// It might be split up into several more specialized clients in the future.
type Client interface {
	// Lists namespaces in the cluster.
	ListNamespaces() (*v1.NamespaceList, error)
	// Lists pods in the specified namespace.
	ListPods(namespace string) (*v1.PodList, error)
	// Fetches information about a single pod, including events and logs.
	GetPod(namespace, id string) (*Pod, error)
	// Delete the pod with the specified name in the specified namespace.
	// Returns the name of the deleted pod.
	DeletePod(namespace, name string) (string, error)
}

// ContextClient defines the interface for working with kubernetes contexts from a view.
type ContextClient interface {
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

// ContextClientImpl is used to manipulate the current context.
type ContextClientImpl struct {
	modifyConfig ModifyConfigFunc
	configAccess clientcmd.ConfigAccess
	config       api.Config
}

// NewContextClientImpl creates a new ContextClient.
//
// If modifyConfig is nil it will default to clientcmd.ModifyConfig.
func NewContextClientImpl(configAccess clientcmd.ConfigAccess, config api.Config, modifyConfig ModifyConfigFunc) *ContextClientImpl {
	impl := &ContextClientImpl{
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
func (c *ContextClientImpl) CurrentApiContext() (*api.Context, bool) {
	ctx, ok := c.config.Contexts[c.config.CurrentContext]
	return ctx, ok
}

// CurrentContext returns the currently active context.
func (c *ContextClientImpl) CurrentContext() string {
	return c.config.CurrentContext
}

// Contexts returns a list of available contexts.
func (c *ContextClientImpl) Contexts() []string {
	return maps.Keys(c.config.Contexts)
}

// SwitchContext changes the active context in a kubeconfig.
func (c *ContextClientImpl) SwitchContext(ctx, namespace string) (err error) {

	kubeCtx, ok := c.config.Contexts[ctx]

	if !ok {
		return fmt.Errorf("context %s doesn't exists", ctx)
	}

	err = c.modifyConfig(c.configAccess, c.config, true)
	if err != nil {
		return fmt.Errorf("error ModifyConfig: %v", err)
	}

	if namespace != "" {
		kubeCtx.Namespace = namespace
	}

	c.config.CurrentContext = ctx

	return nil
}

// DeleteContext deletes a context from a kubeconfig file.
func (c *ContextClientImpl) DeleteContext(ctx string) (err error) {

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
func (c *ContextClientImpl) DeleteUser(user string) (err error) {

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
func (c *ContextClientImpl) DeleteClusterEntry(cluster string) (err error) {

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

// K8sClientImpl is used to fetch data and issue commands to a kubernetes cluster.
type K8sClientImpl struct {
	kubectl kubernetes.Interface
}

// NewK8sClient creates a new K8sClient.
func NewK8sClient(kubectl kubernetes.Interface) Client {
	return &K8sClientImpl{
		kubectl: kubectl,
	}
}

// ListNamespaces fetches all namespaces for the current context.
func (c *K8sClientImpl) ListNamespaces() (*v1.NamespaceList, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	namespaces, err := c.kubectl.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})

	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %v", err)
	}

	return namespaces, nil

}

// ListPods fetches all pods for the current context and namespace.
func (c *K8sClientImpl) ListPods(namespace string) (*v1.PodList, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pods, err := c.kubectl.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})

	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %v", err)
	}

	return pods, nil

}

// GetPod fetches a pod in the current context and namespace.
func (c *K8sClientImpl) GetPod(namespace, id string) (*Pod, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pod, err := c.kubectl.CoreV1().Pods(namespace).Get(ctx, id, metav1.GetOptions{})

	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %v", err)
	}

	eventsCtx, eventsCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer eventsCancel()

	events, err := c.kubectl.CoreV1().Events(namespace).List(eventsCtx, metav1.ListOptions{FieldSelector: fmt.Sprintf("involvedObject.name=%s", pod.Name), TypeMeta: metav1.TypeMeta{Kind: "Pod"}})

	if err != nil {
		return nil, fmt.Errorf("failed to get pod events: %v", err)
	}

	logs := map[string]string{}

	if len(pod.Status.ContainerStatuses) > 0 {
		logs, err = getLogs(c.kubectl, namespace, pod)
	}

	if err != nil {
		return nil, err
	}

	return &Pod{
		Pod:    *pod,
		Events: events.Items,
		Logs:   logs,
	}, nil

}

func getLogs(kubectl kubernetes.Interface, namespace string, pod *v1.Pod) (map[string]string, error) {

	containerLogs := map[string]string{}

	errGroup := &errgroup.Group{}

	for i := range pod.Status.ContainerStatuses {

		container := pod.Status.ContainerStatuses[i]

		errGroup.Go(func() error {

			logsCtx, logsCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer logsCancel()

			tailLines := int64(100)
			logsRequest := kubectl.CoreV1().Pods(namespace).GetLogs(pod.GetName(), &v1.PodLogOptions{Container: container.Name, TailLines: &tailLines})

			podLogs, err := logsRequest.Stream(logsCtx)

			if err != nil {
				return err
			}
			defer podLogs.Close()

			buf := new(bytes.Buffer)
			_, err = io.Copy(buf, podLogs)
			if err != nil {
				return err
			}

			containerLogs[container.Name] = buf.String()

			return nil
		})
	}

	if err := errGroup.Wait(); err != nil {
		return containerLogs, err
	}

	return containerLogs, nil

}

// DeletePod deletes a pod in the current context and namespace.
func (c *K8sClientImpl) DeletePod(namespace, name string) (string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := c.kubectl.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})

	if err != nil {
		return "", fmt.Errorf("failed to delete pod: %v", err)
	}

	return name, nil
}
