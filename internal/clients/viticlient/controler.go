package viticlient

import (
	"context"
	"log/slog"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type controller struct {
	informer cache.SharedIndexInformer
	lister   cache.GenericLister
	queue    workqueue.RateLimitingInterface
	client   dynamic.Interface
	gvr      schema.GroupVersionResource
}

func newController(client dynamic.Interface, gvr schema.GroupVersionResource) *controller {
	factory := dynamicinformer.NewDynamicSharedInformerFactory(client, 30*time.Second)
	informer := factory.ForResource(gvr)

	c := &controller{
		client:   client,
		informer: informer.Informer(),
		lister:   informer.Lister(),
		queue:    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), gvr.Resource),
		gvr:      gvr,
	}

	c.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleAdd,
		UpdateFunc: c.handleUpdate,
		DeleteFunc: c.handleDelete,
	})

	return c
}

func (c *controller) handleAdd(obj any) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err == nil {
		c.queue.Add(key)
	}
}

func (c *controller) handleUpdate(oldObj, newObj any) {
	key, err := cache.MetaNamespaceKeyFunc(newObj)
	if err == nil {
		c.queue.Add(key)
	}
}

func (c *controller) handleDelete(obj any) {
	var key string
	var err error

	if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
		key, err = cache.MetaNamespaceKeyFunc(tombstone.Obj)
	} else {
		key, err = cache.MetaNamespaceKeyFunc(obj)
	}
	if err == nil {
		c.queue.Add(key)
	}
}

func (c *controller) run(ctx context.Context) {
	go c.informer.Run(ctx.Done())

	if !cache.WaitForCacheSync(ctx.Done(), c.informer.HasSynced) {
		panic("failed to sync")
	}

	go wait.Until(c.runWorker, time.Second, ctx.Done())

	ctx.Done()
}

func (c *controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *controller) processNextWorkItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}

	defer c.queue.Done(key)

	if err := c.reconcile(key.(string)); err != nil {
		c.queue.AddRateLimited(key)
		slog.Error("receonciliation of %s failed: %w", key, err)
		return true
	}

	c.queue.Forget(key)
	return true
}

func (c *controller) reconcile(key string) error {

	cachedObj, err := c.lister.Get(key)
	if errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	obj, err := c.informer.get

}
