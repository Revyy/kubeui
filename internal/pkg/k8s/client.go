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

// ContextClient is used to manipulate the current context.
type ContextClient struct {
	configAccess clientcmd.ConfigAccess
	config       api.Config
}

// NewContextClient creates a new ContextClient.
func NewContextClient(configAccess clientcmd.ConfigAccess, config api.Config) *ContextClient {
	return &ContextClient{
		configAccess: configAccess,
		config:       config,
	}
}

// CurrentApiContext returns the currently active api context.
func (c *ContextClient) CurrentApiContext() (*api.Context, bool) {
	ctx, ok := c.config.Contexts[c.config.CurrentContext]
	return ctx, ok
}

// CurrentContext returns the currently active context.
func (c *ContextClient) CurrentContext() string {
	return c.config.CurrentContext
}

// Contexts returns a list of available contexts.
func (c *ContextClient) Contexts() []string {
	return maps.Keys(c.config.Contexts)
}

// SwitchContext changes the active context in a kubeconfig.
func (c *ContextClient) SwitchContext(ctx, namespace string) (err error) {

	kubeCtx, ok := c.config.Contexts[ctx]

	if !ok {
		return fmt.Errorf("context %s doesn't exists", ctx)
	}

	if namespace != "" {
		kubeCtx.Namespace = namespace
	}

	c.config.CurrentContext = ctx
	err = clientcmd.ModifyConfig(c.configAccess, c.config, true)
	if err != nil {
		return fmt.Errorf("error ModifyConfig: %v", err)
	}

	return nil
}

// DeleteContext deletes a context from a kubeconfig file.
func (c *ContextClient) DeleteContext(ctx string) (err error) {

	configFile := c.configAccess.GetDefaultFilename()
	if c.configAccess.IsExplicitFile() {
		configFile = c.configAccess.GetExplicitFile()
	}

	_, ok := c.config.Contexts[ctx]
	if !ok {
		return fmt.Errorf("context %s, is not in %s", ctx, configFile)
	}

	delete(c.config.Contexts, ctx)

	if err := clientcmd.ModifyConfig(c.configAccess, c.config, true); err != nil {
		return err
	}

	return nil
}

// DeleteUser deletes a user entry from a kubeconfig file.
func (c *ContextClient) DeleteUser(user string) (err error) {

	configFile := c.configAccess.GetDefaultFilename()
	if c.configAccess.IsExplicitFile() {
		configFile = c.configAccess.GetExplicitFile()
	}

	_, ok := c.config.AuthInfos[user]
	if !ok {
		return fmt.Errorf("user %s, is not in %s", user, configFile)
	}

	delete(c.config.AuthInfos, user)

	if err := clientcmd.ModifyConfig(c.configAccess, c.config, true); err != nil {
		return err
	}

	return nil
}

// DeleteClusterEntry deletes a cluster entry from a kubeconfig file.
func (c *ContextClient) DeleteClusterEntry(cluster string) (err error) {

	configFile := c.configAccess.GetDefaultFilename()
	if c.configAccess.IsExplicitFile() {
		configFile = c.configAccess.GetExplicitFile()
	}

	_, ok := c.config.Clusters[cluster]
	if !ok {
		return fmt.Errorf("cluster %s, is not in %s", cluster, configFile)
	}

	delete(c.config.Clusters, cluster)

	if err := clientcmd.ModifyConfig(c.configAccess, c.config, true); err != nil {
		return err
	}

	return nil
}

// K8sClient is used to fetch data and issue commands to a kubernetes cluster.
type K8sClient struct {
	kubectl kubernetes.Interface
}

// NewK8sClient creates a new K8sClient.
func NewK8sClient(kubectl kubernetes.Interface) *K8sClient {
	return &K8sClient{
		kubectl: kubectl,
	}
}

// ListNamespaces fetches all namespaces for the current context.
func (c *K8sClient) ListNamespaces() (*v1.NamespaceList, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	namespaces, err := c.kubectl.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})

	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %v", err)
	}

	return namespaces, nil

}

// ListPods fetches all pods for the current context and namespace.
func (c *K8sClient) ListPods(namespace string) (*v1.PodList, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pods, err := c.kubectl.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})

	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %v", err)
	}

	return pods, nil

}

// GetPod fetches a pod in the current context and namespace.
func (c *K8sClient) GetPod(namespace, id string) (*Pod, error) {

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
func (c *K8sClient) DeletePod(namespace, name string) (string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := c.kubectl.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})

	if err != nil {
		return "", fmt.Errorf("failed to delete pod: %v", err)
	}

	return name, nil
}
