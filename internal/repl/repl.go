// Package repl provides the interactive Read-Eval-Print Loop for urload2.
package repl

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/chzyer/readline"

	"github.com/JohnStrunk/URLoad2/internal/downloader"
	"github.com/JohnStrunk/URLoad2/internal/urllist"
)

// DefaultPrompt is the initial prompt displayed by urload2 when awaiting input.
const DefaultPrompt = "urload2 [0]> "

// Version is the current application version string.
const Version = "urload2 v0.1.0-dev"

// Commands returns the list of supported REPL commands.
func Commands() []string {
	return []string{
		"add",
		"clear",
		"exit",
		"get",
		"head",
		"help",
		"list",
		"quit",
		"tail",
		"version",
		"?",
	}
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

// Downloader defines the downloading interface used by REPL.
type Downloader interface {
	DownloadAll(ctx context.Context, urls []string, out io.Writer) error
	TargetName() string
	TargetPath() string
}

// REPL manages the read-eval-print loop session.
type REPL struct {
	in                 io.Reader
	out                io.Writer
	prompt             string
	baseDir            string
	urlList            *urllist.List
	downloader         Downloader
	downloaderExplicit bool
}

// Option configures a REPL instance.
type Option func(*REPL)

// WithPrompt sets a custom prompt string.
func WithPrompt(prompt string) Option {
	return func(r *REPL) {
		r.prompt = prompt
	}
}

// WithURLList sets a custom URL list.
func WithURLList(list *urllist.List) Option {
	return func(r *REPL) {
		r.urlList = list
	}
}

// WithDownloader sets a custom downloader.
func WithDownloader(d Downloader) Option {
	return func(r *REPL) {
		r.downloader = d
		r.downloaderExplicit = true
	}
}

// WithBaseDir sets the working base directory for the downloader.
func WithBaseDir(dir string) Option {
	return func(r *REPL) {
		r.baseDir = dir
	}
}

// New creates a new REPL instance reading from in and writing to out.
func New(in io.Reader, out io.Writer, opts ...Option) *REPL {
	r := &REPL{
		in:  in,
		out: out,
	}
	for _, opt := range opts {
		opt(r)
	}

	if r.urlList == nil {
		r.urlList = urllist.New()
	}

	if r.downloader == nil && !r.downloaderExplicit {
		dlOpts := []downloader.Option{}
		if r.baseDir != "" {
			dlOpts = append(dlOpts, downloader.WithBaseDir(r.baseDir))
		}
		dl, err := downloader.New(dlOpts...)
		if err == nil {
			r.downloader = dl
		}
	}

	return r
}

// currentPrompt returns the prompt string including the current URL list count.
func (r *REPL) currentPrompt() string {
	if r.prompt != "" {
		return r.prompt
	}
	return fmt.Sprintf("urload2 [%d]> ", r.urlList.Len())
}

// URLList returns the underlying urllist.List instance.
func (r *REPL) URLList() *urllist.List {
	return r.urlList
}

// Downloader returns the underlying Downloader instance.
func (r *REPL) Downloader() Downloader {
	return r.downloader
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
		Prompt:       r.currentPrompt(),
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

		shouldExit, err := r.eval(ctx, line)
		if err != nil {
			return err
		}
		if shouldExit {
			return nil
		}

		rl.SetPrompt(r.currentPrompt())
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

		if _, err := fmt.Fprint(r.out, r.currentPrompt()); err != nil {
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

		shouldExit, err := r.eval(ctx, line)
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
func (r *REPL) eval(ctx context.Context, line string) (bool, error) {
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
			"  add <url>  Append a URL to the end of the list\n" +
			"  clear      Clear the list of URLs\n" +
			"  exit       Exit the REPL\n" +
			"  get        Download all URLs in the list to the target directory\n" +
			"  head <n>   Keep the first n URLs in the list\n" +
			"  help, ?    Show available commands\n" +
			"  list       Display the current list of URLs\n" +
			"  quit       Exit the REPL\n" +
			"  tail <n>   Keep the last n URLs in the list\n" +
			"  version    Show version information\n"
		if _, err := fmt.Fprint(r.out, helpText); err != nil {
			return false, fmt.Errorf("failed to write help: %w", err)
		}
		return false, nil

	case "version":
		if _, err := fmt.Fprintln(r.out, Version); err != nil {
			return false, fmt.Errorf("failed to write version: %w", err)
		}
		return false, nil

	case "add":
		if len(fields) < 2 {
			if _, err := fmt.Fprintln(r.out, "error: add requires a URL"); err != nil {
				return false, fmt.Errorf("failed to write add error: %w", err)
			}
			return false, nil
		}
		rawURL := fields[1]
		if err := r.urlList.Add(rawURL); err != nil {
			if _, writeErr := fmt.Fprintf(r.out, "error: %v\n", err); writeErr != nil {
				return false, fmt.Errorf("failed to write add error: %w", writeErr)
			}
			return false, nil
		}
		return false, nil

	case "list":
		urls := r.urlList.Get()
		for _, u := range urls {
			if _, err := fmt.Fprintln(r.out, u); err != nil {
				return false, fmt.Errorf("failed to write list item: %w", err)
			}
		}
		return false, nil

	case "clear":
		r.urlList.Clear()
		return false, nil

	case "head":
		if len(fields) < 2 {
			if _, err := fmt.Fprintln(r.out, "error: head requires a count"); err != nil {
				return false, fmt.Errorf("failed to write head error: %w", err)
			}
			return false, nil
		}
		n, err := strconv.Atoi(fields[1])
		if err != nil || n < 0 {
			if _, writeErr := fmt.Fprintln(r.out, "error: head requires a non-negative integer count"); writeErr != nil {
				return false, fmt.Errorf("failed to write head error: %w", writeErr)
			}
			return false, nil
		}
		if err := r.urlList.Head(n); err != nil {
			if _, writeErr := fmt.Fprintf(r.out, "error: %v\n", err); writeErr != nil {
				return false, fmt.Errorf("failed to write head error: %w", writeErr)
			}
			return false, nil
		}
		return false, nil

	case "tail":
		if len(fields) < 2 {
			if _, err := fmt.Fprintln(r.out, "error: tail requires a count"); err != nil {
				return false, fmt.Errorf("failed to write tail error: %w", err)
			}
			return false, nil
		}
		n, err := strconv.Atoi(fields[1])
		if err != nil || n < 0 {
			if _, writeErr := fmt.Fprintln(r.out, "error: tail requires a non-negative integer count"); writeErr != nil {
				return false, fmt.Errorf("failed to write tail error: %w", writeErr)
			}
			return false, nil
		}
		if err := r.urlList.Tail(n); err != nil {
			if _, writeErr := fmt.Fprintf(r.out, "error: %v\n", err); writeErr != nil {
				return false, fmt.Errorf("failed to write tail error: %w", writeErr)
			}
			return false, nil
		}
		return false, nil

	case "get":
		if r.downloader == nil {
			if _, err := fmt.Fprintln(r.out, "error: downloader not initialized"); err != nil {
				return false, fmt.Errorf("failed to write get error: %w", err)
			}
			return false, nil
		}
		if err := r.downloader.DownloadAll(ctx, r.urlList.Get(), r.out); err != nil {
			if _, writeErr := fmt.Fprintf(r.out, "download error: %v\n", err); writeErr != nil {
				return false, fmt.Errorf("failed to write get error: %w", writeErr)
			}
			return false, nil
		}
		return false, nil

	default:
		if _, err := fmt.Fprintf(r.out, "unknown command: %q. Type 'help' for available commands.\n", command); err != nil {
			return false, fmt.Errorf("failed to write error: %w", err)
		}
		return false, nil
	}
}
