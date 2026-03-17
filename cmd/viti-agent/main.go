package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/NorskHelsenett/ror-viti-agent/internal/clients/viticlient"
	"github.com/NorskHelsenett/ror-viti-agent/internal/config"
	"github.com/NorskHelsenett/ror-viti-agent/internal/converter"
	"github.com/goforj/godump"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {

	conf, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	slog.SetLogLoggerLevel(slog.Level(conf.LogLevel))

	dynamic, err := viticlient.CreateK8sDynamicClient()
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		resources, err := dynamic.Resource(*viticlient.NewGVRV1Alpha1Machine()).List(ctx, metav1.ListOptions{})
		if err != nil {
			slog.ErrorContext(ctx, "failed to gather resource(s)", "error", err)
			time.Sleep(time.Second * time.Duration(conf.PollInterval))
			continue
		}

		slog.InfoContext(ctx, "found resources", "resource_count", len(resources.Items))
		machines, err := viticlient.MarshalMachineObjects(resources.Items)
		if err != nil {
			slog.ErrorContext(ctx, "failed to convert from unstructured", "error", err)
			time.Sleep(time.Second * time.Duration(conf.PollInterval))
			continue
		}
		rormachines, err := converter.ConvertToRorMachines(machines)
		if err != nil {
			slog.ErrorContext(ctx, "failed to convert to ror resources", "error", err)
			time.Sleep(time.Second * 5)
			continue
		}
		godump.DumpJSON(rormachines)

		time.Sleep(time.Second * time.Duration(conf.PollInterval))
	}

}
