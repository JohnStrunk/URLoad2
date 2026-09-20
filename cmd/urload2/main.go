// Package main is the entrypoint for the urload2 application.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/JohnStrunk/URLoad2/internal/repl"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	app := repl.New(os.Stdin, os.Stdout)
	if err := app.Run(ctx); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
