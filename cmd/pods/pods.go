package main

import (
	"flag"
	"kubeui/internal/app/pods"
	"kubeui/internal/pkg/k8s"
	"kubeui/internal/pkg/kubeui"
	"log"
)

func main() {

	// If a specific kubeconfig file is specified then we load that, otherwise the defaults will be loaded.
	kubeconfig := flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()

	clientConfig := k8s.NewClientConfig("", *kubeconfig)

	rawConfig, err := clientConfig.RawConfig()

	if err != nil {
		log.Fatalf("failed to load config")
	}

	configAccess := clientConfig.ConfigAccess()

	clientSet, err := k8s.NewKClientSet(*kubeconfig, configAccess)

	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	m := pods.NewModel(rawConfig, configAccess, clientSet)

	program := kubeui.NewProgram(m, true)
	kubeui.StartProgram(program)

}
