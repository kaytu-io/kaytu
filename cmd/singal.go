package cmd

import (
	"context"
	"os"
	"os/signal"
)

func AppendSignalHandling(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	cancelChan := make(chan os.Signal, 100)

	signal.Notify(cancelChan, os.Interrupt, os.Kill)

	go func() {
		<-cancelChan
		cancel()
		os.Exit(1)
	}()

	return ctx
}
