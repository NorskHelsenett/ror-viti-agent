package viticlient

import (
	"context"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type Event string

const (
	MACHINE_UPDATE Event = "Updated"
	MACHINE_DELETE Event = "Delete"
	MACHINE_CREATE Event = "Create"
)

type MachineEvent struct {
	Type   Event
	Object *unstructured.Unstructured
}

type Watcher interface {
	Watch(ctx context.Context) <-chan MachineEvent
}

func (c *ClientWatcher) Watch1(ctx context.Context, client kubernetes.Interface, gvr schema.GroupVersionResource) (watch.Interface, error) {

	watcher, err := c.dynamicClient.Resource(gvr).Watch(ctx, v1.ListOptions{})
	if err != nil {
		panic(err)
	}

	return watcher, nil
}

func (c *ClientWatcher) Watch2(ctx context.Context, client kubernetes.Interface, gvr schema.GroupVersionResource, synctime time.Duration) (<-chan MachineEvent, error) {

	output := make(chan MachineEvent)
	//
	// factory := informers.NewSharedInformerFactory(client, synctime)
	// informer, err := factory.ForResource(gvr)
	// if err != nil {
	// }
	// watcher, err := c.dynamicClient.Resource(gvr).Watch(ctx, v1.ListOptions{})
	// if err != nil {
	// 	return nil, err
	// }
	// resultChan := watcher.ResultChan()

	return output, nil
}
