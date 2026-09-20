// Package repl provides the interactive Read-Eval-Print Loop for urload2.
package repl

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
)

// DefaultPrompt is the prompt displayed by urload2 when awaiting input.
const DefaultPrompt = "urload2> "

// Version is the current application version string.
const Version = "urload2 v0.1.0-dev"

// Commands returns the list of supported REPL commands.
func Commands() []string {
	return []string{"help", "?", "version", "exit", "quit"}
}

// Complete returns command suggestions matching the given prefix.
func Complete(prefix string) []string {
	var matches []string
	for _, cmd := range Commands() {
		if strings.HasPrefix(cmd, prefix) {
			matches = append(matches, cmd)
		}
	}
	return matches
}

// defaultCompleter creates a readline PrefixCompleter with all supported commands.
func defaultCompleter() readline.AutoCompleter {
	items := make([]readline.PrefixCompleterInterface, 0, len(Commands()))
	for _, cmd := range Commands() {
		items = append(items, readline.PcItem(cmd))
	}
	return readline.NewPrefixCompleter(items...)
}

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

// Complete returns command suggestions matching the given prefix for this REPL instance.
func (r *REPL) Complete(prefix string) []string {
	return Complete(prefix)
}

// Run executes the read-eval-print loop until context cancellation, EOF, or exit command.
func (r *REPL) Run(ctx context.Context) error {
	if file, ok := r.in.(*os.File); ok && readline.IsTerminal(int(file.Fd())) {
		return r.runInteractive(ctx, file)
	}
	return r.runScanner(ctx)
}

// runInteractive runs the REPL using readline for interactive terminals (supports tab completion).
func (r *REPL) runInteractive(ctx context.Context, file *os.File) error {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:       r.prompt,
		AutoComplete: defaultCompleter(),
		Stdin:        file,
		Stdout:       r.out,
	})
	if err != nil {
		return r.runScanner(ctx)
	}
	defer rl.Close()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line, err := rl.Readline()
		if err != nil {
			if err == io.EOF || err == readline.ErrInterrupt {
				return nil
			}
			return fmt.Errorf("read error: %w", err)
		}

		line = strings.TrimSpace(line)
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

// runScanner runs the REPL using a bufio.Scanner for non-interactive or piped inputs.
func (r *REPL) runScanner(ctx context.Context) error {
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

	case "help", "?":
		helpText := "Available commands:\n" +
			"  help, ?  Show available commands\n" +
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
