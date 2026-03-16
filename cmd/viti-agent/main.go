package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/NorskHelsenett/ror-viti-agent/internal/clients/viticlient"
	"github.com/NorskHelsenett/ror-viti-agent/internal/config"
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
		resources, err := dynamic.Resource(*viticlient.NewGVR("stable.example.com", "v1", "crontabs")).List(ctx, metav1.ListOptions{})
		if err != nil {
			slog.ErrorContext(ctx, "failed to gather resource(s)", "error", err)
			time.Sleep(time.Second * 5)
			continue
		}

		slog.InfoContext(ctx, "found resources", "resource_count", len(resources.Items))
		time.Sleep(time.Second * 5)
	}

}
