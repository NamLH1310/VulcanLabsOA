package main

import (
	"context"
	"os"

	"github.com/namlh/vulcanLabsOA/server"
	"github.com/namlh/vulcanLabsOA/util/fmtutil"
)

func main() {
	ctx := context.Background()

	if err := server.Run(ctx, os.Getenv); err != nil {
		fmtutil.Eprintf("%s\n", err)
		os.Exit(1)
	}
}
