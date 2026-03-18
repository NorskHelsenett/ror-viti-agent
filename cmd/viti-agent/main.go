package main

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/NorskHelsenett/ror-viti-agent/internal/clients/rorclient"
	"github.com/NorskHelsenett/ror-viti-agent/internal/clients/viticlient"
	"github.com/NorskHelsenett/ror-viti-agent/internal/config"
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

	slog.InfoContext(ctx, "Creating rorclient")
	_, err = rorclient.NewRorClient(
		ctx,
		conf.RorApikey,
		conf.RorUrl,
		conf.RorRole,
		conf.RorVersion,
		conf.RorCommit,
	)

	slog.InfoContext(ctx, "Creating kubernetes client")
	dynamic, err := viticlient.CreateK8sDynamicClient()
	if err != nil {
		panic(err)
	}

	slog.InfoContext(ctx, "Creating factory")
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynamic, time.Second*5, "", nil)
	slog.InfoContext(ctx, "Creating kubernetes informer")
	informer := factory.ForResource(*viticlient.NewGVRV1Alpha1Machine()).Informer()

	slog.InfoContext(ctx, "Creating handlers")
	handlerfuncs := cache.ResourceEventHandlerFuncs{
		AddFunc:    viticlient.AddFunc,
		UpdateFunc: viticlient.UpdateFunc,
		DeleteFunc: viticlient.DeleteFunc,
	}
	// For ctx stuff, starts and stops with the informer

	slog.InfoContext(ctx, "Registering handlers")
	informer.AddEventHandler(handlerfuncs)

	wg := sync.WaitGroup{}
	wg.Go(func() { informer.RunWithContext(ctx) })

	slog.InfoContext(ctx, "starting cache...")
	if !cache.WaitForNamedCacheSyncWithContext(ctx, informer.HasSynced) {
		panic("failed to sync cache")
	}

	slog.InfoContext(ctx, "cache synced, watching changes")

	wg.Wait()

}
