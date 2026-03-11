package viticlient

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type ClientWatcher struct {
	dynamicClient *dynamic.DynamicClient
	gvr           schema.GroupVersionResource
}

//func newWatcher(client *dynamic.DynamicClient, gvr schema.GroupVersionResource) Watcher {
//	return &clientWatcher{dynamicClient: client, gvr: gvr}
//}

func test2() {
}
