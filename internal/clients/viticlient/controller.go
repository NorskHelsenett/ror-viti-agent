package viticlient

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/NorskHelsenett/ror-viti-agent/internal/clients/rorclient"
	"github.com/NorskHelsenett/ror-viti-agent/internal/converter"
	"github.com/vitistack/common/pkg/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type EventType string

const (
	EVENTADD    EventType = "Add"
	EVENTUPDATE EventType = "Update"
	EVENTDELETE EventType = "Delete"
)

type QueueItem struct {
	Key       string
	Id        string
	EventType EventType
}

type controller struct {
	informer  cache.SharedIndexInformer
	lister    cache.GenericLister
	queue     workqueue.TypedRateLimitingInterface[QueueItem]
	client    dynamic.Interface
	gvr       schema.GroupVersionResource
	rorclient *rorclient.RorClient
	ctx       context.Context
}

func NewController(ctx context.Context, client dynamic.Interface, gvr schema.GroupVersionResource, rorclient *rorclient.RorClient) *controller {
	factory := dynamicinformer.NewDynamicSharedInformerFactory(client, 30*time.Second)
	informer := factory.ForResource(gvr)

	c := &controller{
		client:    client,
		informer:  informer.Informer(),
		lister:    informer.Lister(),
		queue:     workqueue.NewTypedRateLimitingQueue[QueueItem](workqueue.DefaultTypedControllerRateLimiter[QueueItem]()),
		gvr:       gvr,
		rorclient: rorclient,
		ctx:       ctx,
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
	if err != nil {
		slog.Error("could not create key for object: %w", err)
	}
	id := ""
	qItem := QueueItem{
		Key:       key,
		Id:        id,
		EventType: EVENTADD,
	}

	c.queue.Add(qItem)

	slog.Info("add event: adding item to queue", "item", qItem)
}

func (c *controller) handleUpdate(oldObj, newObj any) {
	key, err := cache.MetaNamespaceKeyFunc(newObj)
	if err != nil {
		slog.Error("could not create key for object: %w", err)
	}
	id := ""
	qItem := QueueItem{
		Key:       key,
		Id:        id,
		EventType: EVENTADD,
	}

	c.queue.Add(qItem)
	slog.Info("update event: adding item to queue", "item", qItem)
}

func (c *controller) handleDelete(obj any) {
	var key string
	var err error

	if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
		key, err = cache.MetaNamespaceKeyFunc(tombstone.Obj)
		slog.Info("delete event: adding tombstone key to queue", "key", key)
	} else {
		key, err = cache.MetaNamespaceKeyFunc(obj)
		slog.Info("delete event: adding key to queue", "key", key)
	}

	qItem := QueueItem{
		Key:       key,
		Id:        "",
		EventType: EVENTDELETE,
	}

	if err == nil {
		c.queue.Add(qItem)
	}
}

func (c *controller) Run(ctx context.Context) error {
	go c.informer.Run(ctx.Done())

	if !cache.WaitForCacheSync(ctx.Done(), c.informer.HasSynced) {
		return errors.New("failed to sync cache")
	}

	wg := sync.WaitGroup{}
	wg.Go(func() { wait.Until(c.runWorker, time.Second, ctx.Done()) })

	wg.Wait()
	ctx.Done()
	return nil
}

func (c *controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *controller) processNextWorkItem() bool {
	item, quit := c.queue.Get()
	if quit {
		return false
	}

	defer c.queue.Done(item)

	if err := c.reconcile(item); err != nil {
		c.queue.AddRateLimited(item)
		slog.Error("receonciliation of %s failed: %w", item, err)
		return true
	}

	c.queue.Forget(item)
	return true
}

func (c *controller) reconcile(item QueueItem) error {

	cachedObj, err := c.lister.Get(item.Key)
	if k8serrors.IsNotFound(err) {
		// c.rorclient.DeleteRorResources(c.ctx)
		slog.ErrorContext(c.ctx, "did not find key in cache", "key", item.Key, "error", err)
		return nil
	}
	if err != nil {
		slog.ErrorContext(c.ctx, "unknown error", "key", item.Key, "error", err)
		return err
	}

	var machine v1alpha1.Machine
	err = MarshalAnyMachineObject(cachedObj, &machine)
	if err != nil {
		slog.ErrorContext(c.ctx, "unable to marshal object to machine", "key", item.Key, "error", err)
		return err
	}

	resource, err := converter.ConvertToRorMachine(&machine)
	if err != nil {
		slog.ErrorContext(c.ctx, "unable to convert object to ror resource", "key", item.Key, "error", err)
		return err
	}

	slog.InfoContext(c.ctx, "adding or updating resource", "name", resource.GetName(), "uid", resource.GetUID())
	// err = c.rorclient.UpdateRorResources([]*rorresources.Resource{resource})
	if err != nil {
		return err
	}

	return nil
}
