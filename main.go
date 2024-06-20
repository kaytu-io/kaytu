package main

import (
	"context"
	"github.com/kaytu-io/kaytu/cmd"
)

func main() {
	ctx := cmd.AppendSignalHandling(context.Background())
	cmd.ExecuteContext(ctx)
}
