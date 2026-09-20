// Package main is the entrypoint for the urload2 application.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/JohnStrunk/URLoad2/internal/repl"
)

func run(ctx context.Context, in io.Reader, out, errOut io.Writer) error {
	app := repl.New(in, out)
	if err := app.Run(ctx); err != nil && err != context.Canceled {
		fmt.Fprintf(errOut, "error: %v\n", err)
		return err
	}
	return nil
}

var osExit = os.Exit

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx, os.Stdin, os.Stdout, os.Stderr); err != nil {
		osExit(1)
	}
}
