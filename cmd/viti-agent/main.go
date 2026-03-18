package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/NorskHelsenett/ror-viti-agent/internal/clients/rorclient"
	"github.com/NorskHelsenett/ror-viti-agent/internal/clients/viticlient"
	"github.com/NorskHelsenett/ror-viti-agent/internal/config"
	"github.com/vitistack/common/pkg/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

func main() {

	conf, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	slog.SetLogLoggerLevel(slog.Level(conf.LogLevel))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err = rorclient.NewRorClient(
		ctx,
		conf.RorApikey,
		conf.RorUrl,
		conf.RorRole,
		conf.RorVersion,
		conf.RorCommit,
	)

	dynamic, err := viticlient.CreateK8sDynamicClient()
	if err != nil {
		panic(err)
	}

	// static, err := viticlient.CreateK8sStaticClient(false)

	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynamic, 5, "", nil)
	informer := factory.ForResource(*viticlient.NewGVRV1Alpha1Machine()).Informer()

	informer.RunWithContext(ctx)
	informer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj any) {
				var machine v1alpha1.Machine
				err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.(*unstructured.Unstructured).Object, &machine)
				if err != nil {
					panic(fmt.Errorf("failed to cast to machine"))
				}
				slog.Info("added machine", "name", machine.Name)
			},
			UpdateFunc: func(oldObj, newObj any) {
				oldMachine, ok := oldObj.(v1alpha1.Machine)
				if !ok {
					panic(fmt.Errorf("failed to cast to oldMachine"))
				}

				newMachine, ok := newObj.(v1alpha1.Machine)
				if !ok {
					panic(fmt.Errorf("failed to cast to newMachine"))
				}
				slog.Info("updated machine", "old_name", oldMachine.Name, "new_machine", newMachine.Name)

			},
			DeleteFunc: func(obj any) {
				machine, ok := obj.(v1alpha1.Machine)
				if !ok {
					panic(fmt.Errorf("failed to cast to machine"))
				}
				slog.Info("deleted  machine", "name", machine.Name)

			},
		},
	)

	stopCh := make(chan struct{})
	defer close(stopCh)
	factory.Start(stopCh)

	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		panic("failed to sync cache")
	}

	slog.Info("cache synced, watching changes")
	<-stopCh

	// watcher, err := c.dynamicClient.Resource(gvr).Watch(ctx, v1.ListOptions{})
	// if err != nil {
	// 	return nil, err
	// }
	// resultChan := watcher.ResultChan()
	// watcher, err := dynamic.Resource(*viticlient.NewGVRV1Alpha1Machine()).Watch(ctx, v1.ListOptions{})
	// if err != nil {
	// 	panic(err)
	// }
	// resultChan := watcher.ResultChan()
	// for {
	// 	result := <-resultChan
	// 	godump.DumpJSON(result)
	// }
	// for {
	// 	resources, err := dynamic.Resource(*viticlient.NewGVRV1Alpha1Machine()).List(ctx, metav1.ListOptions{})
	// 	if err != nil {
	// 		slog.ErrorContext(ctx, "failed to gather resource(s)", "error", err)
	// 		time.Sleep(time.Second * time.Duration(conf.PollInterval))
	// 		continue
	// 	}
	//
	// 	slog.InfoContext(ctx, "found resources", "resource_count", len(resources.Items))
	// 	machines, err := viticlient.MarshalMachineObjects(resources.Items)
	// 	if err != nil {
	// 		slog.ErrorContext(ctx, "failed to convert from unstructured", "error", err)
	// 		time.Sleep(time.Second * time.Duration(conf.PollInterval))
	// 		continue
	// 	}
	// 	rormachines, err := converter.ConvertToRorMachines(machines)
	// 	if err != nil {
	// 		slog.ErrorContext(ctx, "failed to convert to ror resources", "error", err)
	// 		time.Sleep(time.Second * 5)
	// 		continue
	// 	}
	//
	// 	rclient.UpdateRorResources(rormachines)
	// 	godump.DumpJSON(rormachines)
	//
	// 	time.Sleep(time.Second * time.Duration(conf.PollInterval))
	// }

}
