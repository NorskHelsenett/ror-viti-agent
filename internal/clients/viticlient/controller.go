package viticlient

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/NorskHelsenett/ror-viti-agent/internal/clients/rorclient"
	"github.com/vitistack/common/pkg/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const RORFinalizer = "ManagedByRoR"

type controller struct {
	informer      cache.SharedIndexInformer
	lister        cache.GenericLister
	queue         workqueue.TypedRateLimitingInterface[string]
	client        dynamic.Interface
	gvr           schema.GroupVersionResource
	machineClient *rorclient.MachineClient
	ctx           context.Context
}

func NewController(ctx context.Context, client dynamic.Interface, gvr schema.GroupVersionResource, machineClient *rorclient.MachineClient) *controller {
	factory := dynamicinformer.NewDynamicSharedInformerFactory(client, 30*time.Second)
	informer := factory.ForResource(gvr)

	c := &controller{
		client:        client,
		informer:      informer.Informer(),
		lister:        informer.Lister(),
		queue:         workqueue.NewTypedRateLimitingQueue[string](workqueue.DefaultTypedControllerRateLimiter[string]()),
		gvr:           gvr,
		machineClient: machineClient,
		ctx:           ctx,
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

	c.queue.Add(key)

	slog.Info("add event: adding key to queue", "key", key)
}

func (c *controller) handleUpdate(oldObj, newObj any) {
	key, err := cache.MetaNamespaceKeyFunc(newObj)
	if err != nil {
		slog.Error("could not create key for object: %w", err)
	}

	c.queue.Add(key)
	slog.Info("update event: adding key to queue", "key", key)
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

	if err == nil {
		c.queue.Add(key)
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

func (c *controller) reconcile(key string) error {

	cachedObj, err := c.lister.Get(key)
	if k8serrors.IsNotFound(err) {
		slog.ErrorContext(c.ctx, "did not find key in cache", "key", key, "error", err)
		return nil
	}
	if err != nil {
		slog.ErrorContext(c.ctx, "unknown error", "key", key, "error", err)
		return err
	}

	var machine v1alpha1.Machine
	err = MarshalAnyMachineObject(cachedObj, &machine)
	if err != nil {
		slog.ErrorContext(c.ctx, "unable to marshal object to machine", "key", key, "error", err)
		return err
	}
	//
	// if the deletetion time stamp is not set we need to add our finalizer, if
	// it is set we need to remove the object from ror and remove our finalizer
	if machine.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(&machine, RORFinalizer) {
			if ok := controllerutil.AddFinalizer(&machine, RORFinalizer); ok {
				slog.Debug("added finalizer", "key", key)
			} else {
				slog.Debug("could not add finalizer", "key", key)
			}

			err := c.UpdateUnstructured(context.TODO(), &machine)
			if err != nil {
				err := fmt.Errorf("could not update resource with key: %s: %w", key, err)
				return err
			}
		}
	} else {
		//delete from ROR
		c.machineClient.Delete(context.TODO(), machine)

		if controllerutil.ContainsFinalizer(&machine, RORFinalizer) {
			if ok := controllerutil.RemoveFinalizer(&machine, RORFinalizer); ok {
				slog.Debug("removed finalizer", "key", key)
			} else {
				slog.Debug("could not remove finalizer", "key", key)
			}

			err := c.UpdateUnstructured(context.TODO(), &machine)
			if err != nil {
				err := fmt.Errorf("could not update resource with key: %s: %w", key, err)
				return err
			}
		} else {
			warn := fmt.Sprintf("resource with delete timestamp did not have the %s finalizer", RORFinalizer)
			slog.Warn(warn, "key", key)
		}

		return nil
	}

	err = c.machineClient.UpdateProviderStatus(context.TODO(), machine)
	if err != nil {
		return err
	}

	return nil
}

func (c controller) UpdateUnstructured(ctx context.Context, machine *v1alpha1.Machine) error {
	// TODO: we should generate a client for the Machine CRD so we don't
	// have to use the dynamic client
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&machine)
	if err != nil {
		return err
	}

	_, err = c.client.Resource(c.gvr).Update(context.TODO(), &unstructured.Unstructured{Object: obj}, v1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}
