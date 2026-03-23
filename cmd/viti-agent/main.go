package main

import (
	"context"
	"log/slog"

	"github.com/NorskHelsenett/ror-viti-agent/internal/clients/rorclient"
	"github.com/NorskHelsenett/ror-viti-agent/internal/clients/viticlient"
	"github.com/NorskHelsenett/ror-viti-agent/internal/config"
)

func main() {
	// TODO: move flag setup here
	// having flags defined where they are used can quickly lead to duplicate
	// definitions.

	conf, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	slog.SetLogLoggerLevel(slog.Level(conf.LogLevel))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slog.InfoContext(ctx, "Creating rorclient")
	rclient, err := rorclient.NewRorClient(
		ctx,
		conf.RorApikey,
		conf.RorUrl,
		conf.RorRole,
		conf.RorVersion,
		conf.RorCommit,
	)
	if err != nil {
		panic(err)
	}

	slog.InfoContext(ctx, "Creating kubernetes client")
	dynamic, err := viticlient.CreateK8sDynamicClient()
	if err != nil {
		panic(err)
	}

	controller := viticlient.NewController(ctx, dynamic, *viticlient.NewGVRV1Alpha1Machine(), rclient)

	err = controller.Run(ctx)
	if err != nil {
		panic(err)
	}

}
