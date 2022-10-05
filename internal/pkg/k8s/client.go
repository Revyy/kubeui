package k8s

import (
	"flag"
	"fmt"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
)

// NewKClientSet creates a kubernetes ClientSet that can be used to issue kubernetes commands.
func NewKClientSet() (*kubernetes.Clientset, error) {

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
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

// SwitchContext changes the active context in a kubeconfig.
func SwitchContext(ctx, namespace string, configAccess clientcmd.ConfigAccess, config api.Config) (err error) {

	kubeCtx, ok := config.Contexts[ctx]

	if !ok {
		return fmt.Errorf("context %s doesn't exists", ctx)
	}

	if namespace != "" {
		kubeCtx.Namespace = namespace
	}

	config.CurrentContext = ctx
	err = clientcmd.ModifyConfig(configAccess, config, true)
	if err != nil {
		return fmt.Errorf("error ModifyConfig: %v", err)
	}

	return nil
}

// DeleteContext deletes a context from a kubeconfig file.
func DeleteContext(ctx string, configAccess clientcmd.ConfigAccess, config api.Config) (err error) {

	configFile := configAccess.GetDefaultFilename()
	if configAccess.IsExplicitFile() {
		configFile = configAccess.GetExplicitFile()
	}

	_, ok := config.Contexts[ctx]
	if !ok {
		return fmt.Errorf("context %s, is not in %s", ctx, configFile)
	}

	delete(config.Contexts, ctx)

	if err := clientcmd.ModifyConfig(configAccess, config, true); err != nil {
		return err
	}

	return nil
}

// DeleteUser deletes a user entry from a kubeconfig file.
func DeleteUser(user string, configAccess clientcmd.ConfigAccess, config api.Config) (err error) {

	configFile := configAccess.GetDefaultFilename()
	if configAccess.IsExplicitFile() {
		configFile = configAccess.GetExplicitFile()
	}

	_, ok := config.AuthInfos[user]
	if !ok {
		return fmt.Errorf("user %s, is not in %s", user, configFile)
	}

	delete(config.AuthInfos, user)

	if err := clientcmd.ModifyConfig(configAccess, config, true); err != nil {
		return err
	}

	return nil
}

// DeleteClusterEntry deletes a cluster entry from a kubeconfig file.
func DeleteClusterEntry(cluster string, configAccess clientcmd.ConfigAccess, config api.Config) (err error) {

	configFile := configAccess.GetDefaultFilename()
	if configAccess.IsExplicitFile() {
		configFile = configAccess.GetExplicitFile()
	}

	_, ok := config.Clusters[cluster]
	if !ok {
		return fmt.Errorf("cluster %s, is not in %s", cluster, configFile)
	}

	delete(config.Clusters, cluster)

	if err := clientcmd.ModifyConfig(configAccess, config, true); err != nil {
		return err
	}

	return nil
}
