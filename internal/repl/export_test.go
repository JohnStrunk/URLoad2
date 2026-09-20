package repl

import (
	"context"
	"os"

	"github.com/chzyer/readline"
)

// DefaultCompleterForTest exposes defaultCompleter for testing.
func DefaultCompleterForTest() readline.AutoCompleter {
	return defaultCompleter()
}

// RunInteractiveForTest exposes runInteractive for testing.
func (r *REPL) RunInteractiveForTest(ctx context.Context, file *os.File) error {
	return r.runInteractive(ctx, file)
}

// EvalForTest exposes eval for testing.
func (r *REPL) EvalForTest(line string) (bool, error) {
	return r.eval(line)
}
