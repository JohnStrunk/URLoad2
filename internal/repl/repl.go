// Package repl provides the interactive Read-Eval-Print Loop for urload2.
package repl

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
)

// DefaultPrompt is the prompt displayed by urload2 when awaiting input.
const DefaultPrompt = "urload2> "

// Version is the current application version string.
const Version = "urload2 v0.1.0-dev"

// REPL manages the read-eval-print loop session.
type REPL struct {
	in     io.Reader
	out    io.Writer
	prompt string
}

// Option configures a REPL instance.
type Option func(*REPL)

// WithPrompt sets a custom prompt string.
func WithPrompt(prompt string) Option {
	return func(r *REPL) {
		r.prompt = prompt
	}
}

// New creates a new REPL instance reading from in and writing to out.
func New(in io.Reader, out io.Writer, opts ...Option) *REPL {
	r := &REPL{
		in:     in,
		out:    out,
		prompt: DefaultPrompt,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Run executes the read-eval-print loop until context cancellation, EOF, or exit command.
func (r *REPL) Run(ctx context.Context) error {
	scanner := bufio.NewScanner(r.in)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if _, err := fmt.Fprint(r.out, r.prompt); err != nil {
			return fmt.Errorf("failed to write prompt: %w", err)
		}

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("read error: %w", err)
			}
			// Reached EOF gracefully
			return nil
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		shouldExit, err := r.eval(line)
		if err != nil {
			return err
		}
		if shouldExit {
			return nil
		}
	}
}

// eval parses and executes a single input command line.
// Returns true if the loop should terminate.
func (r *REPL) eval(line string) (bool, error) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return false, nil
	}

	command := fields[0]

	switch command {
	case "exit", "quit":
		if _, err := fmt.Fprintln(r.out, "Goodbye!"); err != nil {
			return true, fmt.Errorf("failed to write exit message: %w", err)
		}
		return true, nil

	case "help":
		helpText := "Available commands:\n" +
			"  help     Show available commands\n" +
			"  version  Show version information\n" +
			"  exit     Exit the REPL\n" +
			"  quit     Exit the REPL\n"
		if _, err := fmt.Fprint(r.out, helpText); err != nil {
			return false, fmt.Errorf("failed to write help: %w", err)
		}
		return false, nil

	case "version":
		if _, err := fmt.Fprintln(r.out, Version); err != nil {
			return false, fmt.Errorf("failed to write version: %w", err)
		}
		return false, nil

	default:
		if _, err := fmt.Fprintf(r.out, "unknown command: %q. Type 'help' for available commands.\n", command); err != nil {
			return false, fmt.Errorf("failed to write error: %w", err)
		}
		return false, nil
	}
}
