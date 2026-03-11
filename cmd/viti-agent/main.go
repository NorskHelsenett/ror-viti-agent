package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/NorskHelsenett/ror-viti-agent/internal/config"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	_, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	_, err = createk8sDynamicClient()
	if err != nil {
		panic(err)
	}

	//vitiClient := viticlient.NewForConfi
}

func createk8sDynamicClient() (*dynamic.DynamicClient, error) {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"),
			"(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		err := fmt.Errorf("could not build kubeconfig: %w\n", err)
		return nil, err
	}

	dClient, err := dynamic.NewForConfig(config)
	if err != nil {
		err := fmt.Errorf("could not create dynamic client: %w\n", err)
		return nil, err
	}

	return dClient, nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // Windows
}
