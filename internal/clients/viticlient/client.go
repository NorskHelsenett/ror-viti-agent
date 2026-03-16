package viticlient

import (
	"flag"
	"fmt"
	"log/slog"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type ClientWatcher struct {
	dynamicClient *dynamic.DynamicClient
	gvr           schema.GroupVersionResource
}

//func newWatcher(client *dynamic.DynamicClient, gvr schema.GroupVersionResource) Watcher {
//	return &clientWatcher{dynamicClient: client, gvr: gvr}
//}

// CreateK8sDynamicClient will create a dynamic.DynamicClient using the current context.
//
// This client is used to get resources dynamically and custom resources.
//
// Example:
// dynamic, _ := CreateK8sDynamicClient()
// resources, err := dynamic.Resource(*viticlient.NewGroupVersionResource("stable.example.com", "v1", "crontabs")).List(ctx, metav1.ListOptions{})
func CreateK8sDynamicClient() (*dynamic.DynamicClient, error) {
	config, err := buildConfig()
	if err != nil {
		slog.Error("Error getting out of cluster config", "error", err)
		return nil, err
	}

	dClient, err := dynamic.NewForConfig(config)
	if err != nil {
		err := fmt.Errorf("could not create dynamic client: %w\n", err)
		return nil, err
	}

	return dClient, nil
}

// CreateK8sStaticClient will create a kubernetes.Clientset based on the inCluster parameter.
// If inCluster is set to "true" it will gather the context from the user.
// If inCluster is set to "false" it expects it's inside the cluster and gathers the config directly.
//
// This client is used to get standard kuberentes resources (pods, namespaces, etc.).
//
// Example - list all pods in the "default" namespace:
// client, _ := CreateK8sStaticClient(false)
// pods, _ := clientset.CoreV1().Pods("default").List(ctx, metav1.ListOptions{})
func CreateK8sStaticClient(inCluster bool) (*kubernetes.Clientset, error) {
	var clientset *kubernetes.Clientset
	var err error
	slog.Debug("gathering cluster config", "inCluster", inCluster)
	if inCluster {
		clientset, err = getOutClusterClientSet()
		if err != nil {
			return nil, err
		}
	} else {
		clientset, err = getInClusterClientSet()
		if err != nil {
			return nil, err
		}
	}

	return clientset, nil
}

func getInClusterClientSet() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		slog.Error("Error getting in cluster config", "error", err)
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		slog.Error("Error getting clientset", "error", err)
		return nil, err
	}

	return clientset, err

}

func getOutClusterClientSet() (*kubernetes.Clientset, error) {
	config, err := buildConfig()
	if err != nil {
		slog.Error("Error getting out of cluster config", "error", err)
		return nil, err
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		slog.Error("Error getting clientset", "error", err)
		return nil, err
	}
	return clientset, err
}

func buildConfig() (*rest.Config, error) {
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
	return config, nil
}
