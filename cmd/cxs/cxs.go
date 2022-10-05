package main

import (
	"flag"
	"kubeui/internal/app/cxs"
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

	m := cxs.NewModel(rawConfig, configAccess)

	program := kubeui.NewProgram(m)
	kubeui.StartProgram(program)

}
