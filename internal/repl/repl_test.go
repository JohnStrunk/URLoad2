package repl_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/JohnStrunk/URLoad2/internal/repl"
)

func TestREPLExit(t *testing.T) {
	input := "exit\n"
	var out bytes.Buffer

	r := repl.New(strings.NewReader(input), &out)
	err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	output := out.String()
	if !strings.Contains(output, repl.DefaultPrompt) {
		t.Errorf("expected prompt in output, got %q", output)
	}
	if !strings.Contains(output, "Goodbye!") {
		t.Errorf("expected Goodbye! in output, got %q", output)
	}
}

func TestREPLQuit(t *testing.T) {
	input := "quit\n"
	var out bytes.Buffer

	r := repl.New(strings.NewReader(input), &out)
	err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Goodbye!") {
		t.Errorf("expected Goodbye! in output, got %q", output)
	}
}

func TestREPLHelp(t *testing.T) {
	input := "help\nexit\n"
	var out bytes.Buffer

	r := repl.New(strings.NewReader(input), &out)
	err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Available commands:") {
		t.Errorf("expected help output, got %q", output)
	}
	if !strings.Contains(output, "help, ?") {
		t.Errorf("expected 'help, ?' in output, got %q", output)
	}
}

func TestREPLQuestionMarkAlias(t *testing.T) {
	input := "?\nexit\n"
	var out bytes.Buffer

	r := repl.New(strings.NewReader(input), &out)
	err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Available commands:") {
		t.Errorf("expected help output for '?', got %q", output)
	}
	if !strings.Contains(output, "help, ?") {
		t.Errorf("expected 'help, ?' in output, got %q", output)
	}
}

func TestREPLVersion(t *testing.T) {
	input := "version\nexit\n"
	var out bytes.Buffer

	r := repl.New(strings.NewReader(input), &out)
	err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	output := out.String()
	if !strings.Contains(output, repl.Version) {
		t.Errorf("expected version output, got %q", output)
	}
}

func TestREPLUnknownCommand(t *testing.T) {
	input := "foobar\nexit\n"
	var out bytes.Buffer

	r := repl.New(strings.NewReader(input), &out)
	err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	output := out.String()
	if !strings.Contains(output, `unknown command: "foobar"`) {
		t.Errorf("expected unknown command error, got %q", output)
	}
}

func TestREPLEOF(t *testing.T) {
	input := ""
	var out bytes.Buffer

	r := repl.New(strings.NewReader(input), &out)
	err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("expected nil error on EOF, got %v", err)
	}

	output := out.String()
	if output != repl.DefaultPrompt {
		t.Errorf("expected only prompt on EOF, got %q", output)
	}
}

func TestREPLCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	input := "help\n"
	var out bytes.Buffer

	r := repl.New(strings.NewReader(input), &out)
	err := r.Run(ctx)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestREPLCompletion(t *testing.T) {
	tests := []struct {
		prefix   string
		expected []string
	}{
		{
			prefix:   "he",
			expected: []string{"help"},
		},
		{
			prefix:   "?",
			expected: []string{"?"},
		},
		{
			prefix:   "v",
			expected: []string{"version"},
		},
		{
			prefix:   "nonexistent",
			expected: nil,
		},
		{
			prefix:   "",
			expected: []string{"help", "?", "version", "exit", "quit"},
		},
	}

	for _, tt := range tests {
		matches := repl.Complete(tt.prefix)
		if len(matches) != len(tt.expected) {
			t.Fatalf("for prefix %q, expected %d matches, got %d: %v", tt.prefix, len(tt.expected), len(matches), matches)
		}
		for i, exp := range tt.expected {
			if matches[i] != exp {
				t.Errorf("for prefix %q index %d, expected %q, got %q", tt.prefix, i, exp, matches[i])
			}
		}
	}
}

func TestREPLWithCustomPrompt(t *testing.T) {
	input := "exit\n"
	var out bytes.Buffer

	prompt := "custom-prompt> "
	r := repl.New(strings.NewReader(input), &out, repl.WithPrompt(prompt))
	err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	output := out.String()
	if !strings.Contains(output, prompt) {
		t.Errorf("expected custom prompt %q in output, got %q", prompt, output)
	}
}

func TestREPLInstanceComplete(t *testing.T) {
	var in, out bytes.Buffer
	r := repl.New(&in, &out)

	matches := r.Complete("qui")
	if len(matches) != 1 || matches[0] != "quit" {
		t.Errorf("expected ['quit'], got %v", matches)
	}
}

func TestREPLDefaultCompleter(t *testing.T) {
	completer := repl.DefaultCompleterForTest()
	if completer == nil {
		t.Fatal("expected non-nil default completer")
	}

	newLine, length := completer.Do([]rune("he"), 2)
	if length != 2 || len(newLine) == 0 {
		t.Fatalf("expected completer to suggest completion for 'he', got len=%d, newLine=%v", length, newLine)
	}
}

func TestREPLEmptyAndWhitespaceLines(t *testing.T) {
	input := "\n   \n\t  \r\nexit\n"
	var out bytes.Buffer

	r := repl.New(strings.NewReader(input), &out)
	err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Goodbye!") {
		t.Errorf("expected Goodbye! in output, got %q", output)
	}
}

type errorReader struct {
	err error
}

func (r *errorReader) Read(_ []byte) (int, error) {
	return 0, r.err
}

func TestREPLScannerError(t *testing.T) {
	expectedErr := errors.New("read broken")
	r := repl.New(&errorReader{err: expectedErr}, &bytes.Buffer{})

	err := r.Run(context.Background())
	if err == nil || !strings.Contains(err.Error(), "read error: read broken") {
		t.Fatalf("expected read error, got %v", err)
	}
}

type failWriter struct {
	failAfter int
	written   int
	err       error
}

func (f *failWriter) Write(p []byte) (int, error) {
	if f.written+len(p) > f.failAfter {
		return 0, f.err
	}
	f.written += len(p)
	return len(p), nil
}

func TestREPLWritePromptError(t *testing.T) {
	writeErr := errors.New("write failed")
	w := &failWriter{failAfter: 0, err: writeErr}
	r := repl.New(strings.NewReader("exit\n"), w)

	err := r.Run(context.Background())
	if err == nil || !strings.Contains(err.Error(), "failed to write prompt") {
		t.Fatalf("expected failed to write prompt error, got %v", err)
	}
}

func TestREPLEvalWriteErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		errContains string
	}{
		{
			name:        "exit write error",
			input:       "exit\n",
			errContains: "failed to write exit message",
		},
		{
			name:        "help write error",
			input:       "help\n",
			errContains: "failed to write help",
		},
		{
			name:        "version write error",
			input:       "version\n",
			errContains: "failed to write version",
		},
		{
			name:        "unknown command write error",
			input:       "unknown_cmd\n",
			errContains: "failed to write error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writeErr := errors.New("cannot write")
			// Allow writing the prompt ("urload2> "), but fail writing the command response
			w := &failWriter{failAfter: len(repl.DefaultPrompt), err: writeErr}
			r := repl.New(strings.NewReader(tt.input), w)

			err := r.Run(context.Background())
			if err == nil || !strings.Contains(err.Error(), tt.errContains) {
				t.Fatalf("expected %q error, got %v", tt.errContains, err)
			}
		})
	}
}

func TestRunInteractive(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "repl-interactive-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString("   help   \n\n   exit   \n"); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("failed to seek temp file: %v", err)
	}

	var out bytes.Buffer
	r := repl.New(tmpFile, &out)
	err = r.RunInteractiveForTest(context.Background(), tmpFile)
	if err != nil {
		t.Fatalf("expected nil error from RunInteractive, got %v", err)
	}

	if !strings.Contains(out.String(), "Available commands:") {
		t.Errorf("expected help output in interactive mode, got %q", out.String())
	}
	if !strings.Contains(out.String(), "Goodbye!") {
		t.Errorf("expected Goodbye! in interactive mode, got %q", out.String())
	}
}

func TestRunInteractiveEOF(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "repl-interactive-eof-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	var out bytes.Buffer
	r := repl.New(tmpFile, &out)
	err = r.RunInteractiveForTest(context.Background(), tmpFile)
	if err != nil {
		t.Fatalf("expected nil error on EOF in interactive mode, got %v", err)
	}
}

func TestREPLEvalEmpty(t *testing.T) {
	r := repl.New(&bytes.Buffer{}, &bytes.Buffer{})
	exit, err := r.EvalForTest("   ")
	if err != nil || exit {
		t.Fatalf("expected false, nil for whitespace eval, got %v, %v", exit, err)
	}
}

func TestREPLRunFileNonTerminal(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "repl-nonterminal-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString("exit\n"); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("failed to seek temp file: %v", err)
	}

	var out bytes.Buffer
	r := repl.New(tmpFile, &out)
	err = r.Run(context.Background())
	if err != nil {
		t.Fatalf("expected nil error from Run with non-terminal file, got %v", err)
	}
	if !strings.Contains(out.String(), "Goodbye!") {
		t.Errorf("expected Goodbye! in output, got %q", out.String())
	}
}
