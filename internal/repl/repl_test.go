package repl_test

import (
	"bytes"
	"context"
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
