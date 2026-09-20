package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

type failingReader struct{}

func (f *failingReader) Read(_ []byte) (int, error) {
	return 0, errors.New("read failed")
}

func TestRunSuccess(t *testing.T) {
	in := strings.NewReader("exit\n")
	var out, errOut bytes.Buffer

	err := run(context.Background(), in, &out, &errOut)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !strings.Contains(out.String(), "Goodbye!") {
		t.Errorf("expected Goodbye! in output, got %q", out.String())
	}
	if errOut.Len() != 0 {
		t.Errorf("expected empty errOut, got %q", errOut.String())
	}
}

func TestRunCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	in := strings.NewReader("help\n")
	var out, errOut bytes.Buffer

	err := run(ctx, in, &out, &errOut)
	if err != nil {
		t.Fatalf("expected nil error on canceled context, got %v", err)
	}
	if errOut.Len() != 0 {
		t.Errorf("expected empty errOut, got %q", errOut.String())
	}
}

func TestRunError(t *testing.T) {
	in := &failingReader{}
	var out, errOut bytes.Buffer

	err := run(context.Background(), in, &out, &errOut)
	if err == nil {
		t.Fatal("expected error from failingReader, got nil")
	}

	if !strings.Contains(errOut.String(), "error: ") {
		t.Errorf("expected error message in errOut, got %q", errOut.String())
	}
}

func TestMainSuccess(t *testing.T) {
	origExit := osExit
	origStdin := os.Stdin
	defer func() {
		osExit = origExit
		os.Stdin = origStdin
	}()

	var exitCalled bool
	osExit = func(_ int) {
		exitCalled = true
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdin = r
	if _, err := w.WriteString("exit\n"); err != nil {
		t.Fatalf("failed to write pipe: %v", err)
	}
	_ = w.Close()

	main()

	if exitCalled {
		t.Error("expected osExit not to be called on clean exit")
	}
}

func TestMainError(t *testing.T) {
	origExit := osExit
	origStdin := os.Stdin
	defer func() {
		osExit = origExit
		os.Stdin = origStdin
	}()

	var exitCode int
	osExit = func(code int) {
		exitCode = code
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	// Close read pipe immediately so read returns error
	_ = r.Close()
	_ = w.Close()
	os.Stdin = r

	main()

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}
