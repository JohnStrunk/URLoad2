package repl_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/JohnStrunk/URLoad2/internal/repl"
	"github.com/JohnStrunk/URLoad2/internal/urllist"
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
			expected: []string{"head", "help"},
		},
		{
			prefix:   "h",
			expected: []string{"head", "help", "href"},
		},
		{
			prefix:   "i",
			expected: []string{"img"},
		},
		{
			prefix:   "s",
			expected: []string{"sort"},
		},
		{
			prefix:   "u",
			expected: []string{"uniq"},
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
			expected: repl.Commands(),
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
		{
			name:        "add missing url write error",
			input:       "add\n",
			errContains: "failed to write add error",
		},
		{
			name:        "add invalid url write error",
			input:       "add invalid\n",
			errContains: "failed to write add error",
		},
		{
			name:        "head missing count write error",
			input:       "head\n",
			errContains: "failed to write head error",
		},
		{
			name:        "tail missing count write error",
			input:       "tail\n",
			errContains: "failed to write tail error",
		},
		{
			name:        "get nil downloader write error",
			input:       "get\n",
			errContains: "failed to write get error",
		},
		{
			name:        "get download error write error",
			input:       "get\n",
			errContains: "failed to write get error",
		},
		{
			name:        "list write error",
			input:       "list\n",
			errContains: "failed to write list item",
		},
		{
			name:        "href write error",
			input:       "href\n",
			errContains: "failed to write href error",
		},
		{
			name:        "img write error",
			input:       "img\n",
			errContains: "failed to write img error",
		},
		{
			name:        "href http error write error",
			input:       "href\n",
			errContains: "failed to write href output",
		},
		{
			name:        "img http error write error",
			input:       "img\n",
			errContains: "failed to write img output",
		},
		{
			name:        "href success write error",
			input:       "href\n",
			errContains: "failed to write href output",
		},
		{
			name:        "img success write error",
			input:       "img\n",
			errContains: "failed to write img output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writeErr := errors.New("cannot write")
			w := &failWriter{failAfter: len(repl.DefaultPrompt), err: writeErr}
			var r *repl.REPL
			switch tt.name {
			case "get nil downloader write error":
				r = repl.New(strings.NewReader(tt.input), w, repl.WithDownloader(nil))
			case "get download error write error":
				mockDL := &mockDownloader{err: errors.New("download failed")}
				r = repl.New(strings.NewReader(tt.input), w, repl.WithDownloader(mockDL))
				_ = r.URLList().Add("http://example.com/item")
			case "list write error":
				r = repl.New(strings.NewReader(tt.input), w)
				_ = r.URLList().Add("http://example.com/item")
			case "href write error":
				mockHTTP := &mockHTTPClient{
					doFunc: func(_ *http.Request) (*http.Response, error) {
						return nil, errors.New("network failure")
					},
				}
				r = repl.New(strings.NewReader(tt.input), w, repl.WithHTTPClient(mockHTTP))
				_ = r.URLList().Add("http://example.com/item")
			case "img write error":
				mockHTTP := &mockHTTPClient{
					doFunc: func(_ *http.Request) (*http.Response, error) {
						return nil, errors.New("network failure")
					},
				}
				r = repl.New(strings.NewReader(tt.input), w, repl.WithHTTPClient(mockHTTP))
				_ = r.URLList().Add("http://example.com/item")
			case "href http error write error":
				mockHTTP := &mockHTTPClient{
					doFunc: func(_ *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusNotFound,
							Body:       io.NopCloser(strings.NewReader("404")),
						}, nil
					},
				}
				r = repl.New(strings.NewReader(tt.input), w, repl.WithHTTPClient(mockHTTP))
				_ = r.URLList().Add("http://example.com/item")
			case "img http error write error":
				mockHTTP := &mockHTTPClient{
					doFunc: func(_ *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusNotFound,
							Body:       io.NopCloser(strings.NewReader("404")),
						}, nil
					},
				}
				r = repl.New(strings.NewReader(tt.input), w, repl.WithHTTPClient(mockHTTP))
				_ = r.URLList().Add("http://example.com/item")
			case "href success write error":
				mockHTTP := &mockHTTPClient{
					doFunc: func(_ *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(strings.NewReader("<a href='/link'>Link</a>")),
						}, nil
					},
				}
				r = repl.New(strings.NewReader(tt.input), w, repl.WithHTTPClient(mockHTTP))
				_ = r.URLList().Add("http://example.com/item")
			case "img success write error":
				mockHTTP := &mockHTTPClient{
					doFunc: func(_ *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(strings.NewReader("<img src='/img.png'>")),
						}, nil
					},
				}
				r = repl.New(strings.NewReader(tt.input), w, repl.WithHTTPClient(mockHTTP))
				_ = r.URLList().Add("http://example.com/item")
			default:
				r = repl.New(strings.NewReader(tt.input), w)
			}

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

func TestREPLPromptUpdatesWithCount(t *testing.T) {
	input := "add http://example.com/1\nadd http://example.com/2\nclear\nexit\n"
	var out bytes.Buffer

	r := repl.New(strings.NewReader(input), &out)
	err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "URLoad2 [0000] (0)> ") {
		t.Errorf("expected URLoad2 [0000] (0)> in output, got %q", output)
	}
	if !strings.Contains(output, "URLoad2 [0000] (1)> ") {
		t.Errorf("expected URLoad2 [0000] (1)> in output, got %q", output)
	}
	if !strings.Contains(output, "URLoad2 [0000] (2)> ") {
		t.Errorf("expected URLoad2 [0000] (2)> in output, got %q", output)
	}
}

func TestREPLPromptIncludesTargetDirectory(t *testing.T) {
	mockDL := &mockDownloader{targetName: "0007", targetPath: "/tmp/0007"}
	var out bytes.Buffer

	r := repl.New(strings.NewReader("exit\n"), &out, repl.WithDownloader(mockDL))
	err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "URLoad2 [0007] (0)> ") {
		t.Errorf("expected URLoad2 [0007] (0)> in output, got %q", output)
	}
}

func TestREPLAddCommands(t *testing.T) {
	t.Run("missing url", func(t *testing.T) {
		var out bytes.Buffer
		r := repl.New(strings.NewReader("add\nexit\n"), &out)
		err := r.Run(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out.String(), "error: add requires a URL") {
			t.Errorf("expected error message in output, got %q", out.String())
		}
	})

	t.Run("invalid url", func(t *testing.T) {
		var out bytes.Buffer
		r := repl.New(strings.NewReader("add not-a-valid-url\nexit\n"), &out)
		err := r.Run(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out.String(), "error:") {
			t.Errorf("expected error message in output, got %q", out.String())
		}
		if r.URLList().Len() != 0 {
			t.Errorf("expected list to remain empty, got %d", r.URLList().Len())
		}
	})
}

func TestREPLList(t *testing.T) {
	input := "add http://example.com/1\nadd http://example.com/2\nlist\nexit\n"
	var out bytes.Buffer

	r := repl.New(strings.NewReader(input), &out)
	err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "http://example.com/1\n") {
		t.Errorf("expected http://example.com/1 in list output, got %q", output)
	}
	if !strings.Contains(output, "http://example.com/2\n") {
		t.Errorf("expected http://example.com/2 in list output, got %q", output)
	}
}

func TestREPLClear(t *testing.T) {
	input := "add http://example.com/1\nclear\nexit\n"
	var out bytes.Buffer

	r := repl.New(strings.NewReader(input), &out)
	err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r.URLList().Len() != 0 {
		t.Errorf("expected list to be empty after clear, got len %d", r.URLList().Len())
	}
}

func TestREPLHead(t *testing.T) {
	t.Run("valid head", func(t *testing.T) {
		input := "add http://example.com/1\nadd http://example.com/2\nadd http://example.com/3\nhead 1\nexit\n"
		var out bytes.Buffer

		r := repl.New(strings.NewReader(input), &out)
		err := r.Run(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if r.URLList().Len() != 1 {
			t.Fatalf("expected len 1, got %d", r.URLList().Len())
		}
		if r.URLList().Get()[0] != "http://example.com/1" {
			t.Errorf("expected http://example.com/1, got %q", r.URLList().Get()[0])
		}
	})

	t.Run("missing count", func(t *testing.T) {
		var out bytes.Buffer
		r := repl.New(strings.NewReader("head\nexit\n"), &out)
		err := r.Run(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out.String(), "error: head requires a count") {
			t.Errorf("expected head count error, got %q", out.String())
		}
	})

	t.Run("invalid count", func(t *testing.T) {
		var out bytes.Buffer
		r := repl.New(strings.NewReader("head abc\nexit\n"), &out)
		err := r.Run(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out.String(), "error: head requires a non-negative integer count") {
			t.Errorf("expected invalid count error, got %q", out.String())
		}
	})
}

func TestREPLTail(t *testing.T) {
	t.Run("valid tail", func(t *testing.T) {
		input := "add http://example.com/1\nadd http://example.com/2\nadd http://example.com/3\ntail 1\nexit\n"
		var out bytes.Buffer

		r := repl.New(strings.NewReader(input), &out)
		err := r.Run(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if r.URLList().Len() != 1 {
			t.Fatalf("expected len 1, got %d", r.URLList().Len())
		}
		if r.URLList().Get()[0] != "http://example.com/3" {
			t.Errorf("expected http://example.com/3, got %q", r.URLList().Get()[0])
		}
	})

	t.Run("missing count", func(t *testing.T) {
		var out bytes.Buffer
		r := repl.New(strings.NewReader("tail\nexit\n"), &out)
		err := r.Run(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out.String(), "error: tail requires a count") {
			t.Errorf("expected tail count error, got %q", out.String())
		}
	})

	t.Run("invalid count", func(t *testing.T) {
		var out bytes.Buffer
		r := repl.New(strings.NewReader("tail -5\nexit\n"), &out)
		err := r.Run(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out.String(), "error: tail requires a non-negative integer count") {
			t.Errorf("expected invalid count error, got %q", out.String())
		}
	})
}

type mockDownloader struct {
	downloadedURLs []string
	err            error
	targetName     string
	targetPath     string
}

func (m *mockDownloader) DownloadAll(_ context.Context, urls []string, out io.Writer) error {
	m.downloadedURLs = append(m.downloadedURLs, urls...)
	if m.err != nil {
		return m.err
	}
	fmt.Fprintf(out, "Downloaded %d URLs\n", len(urls))
	return nil
}

func (m *mockDownloader) TargetName() string { return m.targetName }
func (m *mockDownloader) TargetPath() string { return m.targetPath }

func TestREPLGet(t *testing.T) {
	mockDL := &mockDownloader{targetName: "0000", targetPath: "/tmp/0000"}
	input := "add http://example.com/test\nget\nexit\n"
	var out bytes.Buffer

	r := repl.New(strings.NewReader(input), &out, repl.WithDownloader(mockDL))
	err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mockDL.downloadedURLs) != 1 || mockDL.downloadedURLs[0] != "http://example.com/test" {
		t.Errorf("mock downloader did not receive expected URLs, got %v", mockDL.downloadedURLs)
	}
	if !strings.Contains(out.String(), "Downloaded 1 URLs") {
		t.Errorf("expected output to contain Downloaded 1 URLs, got %q", out.String())
	}
}

func TestREPLGetError(t *testing.T) {
	mockDL := &mockDownloader{err: errors.New("mock download failure")}
	input := "add http://example.com/test\nget\nexit\n"
	var out bytes.Buffer

	r := repl.New(strings.NewReader(input), &out, repl.WithDownloader(mockDL))
	err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "download error: mock download failure") {
		t.Errorf("expected download error in output, got %q", out.String())
	}
}

func TestREPLGetNilDownloader(t *testing.T) {
	var out bytes.Buffer
	r := repl.New(strings.NewReader("get\nexit\n"), &out, repl.WithDownloader(nil))
	err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "error: downloader not initialized") {
		t.Errorf("expected nil downloader error, got %q", out.String())
	}
}

func TestREPLOptions(t *testing.T) {
	tempDir := t.TempDir()
	customList := urllist.New()
	_ = customList.Add("http://example.com/custom")
	mockHTTP := &mockHTTPClient{}
	r := repl.New(strings.NewReader("exit\n"), &bytes.Buffer{},
		repl.WithBaseDir(tempDir),
		repl.WithURLList(customList),
		repl.WithHTTPClient(mockHTTP),
	)
	if r.Downloader() == nil {
		t.Errorf("expected non-nil downloader with baseDir")
	}
	if r.URLList() != customList {
		t.Errorf("expected custom URLList")
	}
}

type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.doFunc != nil {
		return m.doFunc(req)
	}
	return nil, errors.New("mock http client unhandled request")
}

func TestREPLSort(t *testing.T) {
	input := "add http://example.com/charlie\nadd http://example.com/alpha\nadd http://example.com/bravo\nsort\nlist\nexit\n"
	var out bytes.Buffer

	r := repl.New(strings.NewReader(input), &out)
	err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		"http://example.com/alpha",
		"http://example.com/bravo",
		"http://example.com/charlie",
	}
	urls := r.URLList().Get()
	if len(urls) != len(expected) {
		t.Fatalf("expected %d urls, got %d: %v", len(expected), len(urls), urls)
	}
	for i, u := range urls {
		if u != expected[i] {
			t.Errorf("at index %d expected %q, got %q", i, expected[i], u)
		}
	}
}

func TestREPLUniq(t *testing.T) {
	input := "add http://example.com/a\nadd http://example.com/b\nadd http://example.com/a\nadd http://example.com/c\nadd http://example.com/b\nuniq\nlist\nexit\n"
	var out bytes.Buffer

	r := repl.New(strings.NewReader(input), &out)
	err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		"http://example.com/a",
		"http://example.com/b",
		"http://example.com/c",
	}
	urls := r.URLList().Get()
	if len(urls) != len(expected) {
		t.Fatalf("expected %d urls, got %d: %v", len(expected), len(urls), urls)
	}
	for i, u := range urls {
		if u != expected[i] {
			t.Errorf("at index %d expected %q, got %q", i, expected[i], u)
		}
	}
}

func TestREPLHref(t *testing.T) {
	htmlContent := `<!DOCTYPE html><html><body>
		<a href="/about">About</a>
		<a href="https://other.com/page">External</a>
		<a href="mailto:test@example.com">Email</a>
	</body></html>`

	mockClient := &mockHTTPClient{
		doFunc: func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(htmlContent)),
			}, nil
		},
	}

	input := "add http://example.com/root\nhref\nexit\n"
	var out bytes.Buffer

	r := repl.New(strings.NewReader(input), &out, repl.WithHTTPClient(mockClient))
	err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	urls := r.URLList().Get()
	expected := []string{
		"http://example.com/about",
		"https://other.com/page",
	}
	if len(urls) != len(expected) {
		t.Fatalf("expected %d urls, got %d: %v", len(expected), len(urls), urls)
	}
	for i, u := range urls {
		if u != expected[i] {
			t.Errorf("at index %d expected %q, got %q", i, expected[i], u)
		}
	}
	if !strings.Contains(out.String(), "Scanning http://example.com/root => 200 (found 2)") {
		t.Errorf("expected scanning output in %q", out.String())
	}
}

func TestREPLHrefError(t *testing.T) {
	mockClient := &mockHTTPClient{
		doFunc: func(_ *http.Request) (*http.Response, error) {
			return nil, errors.New("connection refused")
		},
	}

	input := "add http://example.com/bad\nhref\nexit\n"
	var out bytes.Buffer

	r := repl.New(strings.NewReader(input), &out, repl.WithHTTPClient(mockClient))
	err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "Scanning http://example.com/bad => error:") || !strings.Contains(out.String(), "connection refused") {
		t.Errorf("expected error message in output, got %q", out.String())
	}
	if r.URLList().Len() != 0 {
		t.Errorf("expected url list to be emptied after href extraction failure, got %d items", r.URLList().Len())
	}
}

func TestREPLHrefHTTPError(t *testing.T) {
	mockClient := &mockHTTPClient{
		doFunc: func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("404 page not found")),
			}, nil
		},
	}

	input := "add http://example.com/404\nhref\nexit\n"
	var out bytes.Buffer

	r := repl.New(strings.NewReader(input), &out, repl.WithHTTPClient(mockClient))
	err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "Scanning http://example.com/404 => 404") {
		t.Errorf("expected 404 scanning message in output, got %q", out.String())
	}
	if r.URLList().Len() != 0 {
		t.Errorf("expected url list to be emptied after 404, got %d items", r.URLList().Len())
	}
}

func TestREPLImg(t *testing.T) {
	htmlContent := `<!DOCTYPE html><html><body>
		<img src="/images/pic.png" />
		<img src="https://cdn.example.com/banner.jpg" />
		<img src="data:image/png;base64,iVBOR" />
	</body></html>`

	mockClient := &mockHTTPClient{
		doFunc: func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(htmlContent)),
			}, nil
		},
	}

	input := "add http://example.com/gallery\nimg\nexit\n"
	var out bytes.Buffer

	r := repl.New(strings.NewReader(input), &out, repl.WithHTTPClient(mockClient))
	err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	urls := r.URLList().Get()
	expected := []string{
		"http://example.com/images/pic.png",
		"https://cdn.example.com/banner.jpg",
	}
	if len(urls) != len(expected) {
		t.Fatalf("expected %d urls, got %d: %v", len(expected), len(urls), urls)
	}
	for i, u := range urls {
		if u != expected[i] {
			t.Errorf("at index %d expected %q, got %q", i, expected[i], u)
		}
	}
	if !strings.Contains(out.String(), "Scanning http://example.com/gallery => 200 (found 2)") {
		t.Errorf("expected scanning output in %q", out.String())
	}
}

func TestREPLImgError(t *testing.T) {
	mockClient := &mockHTTPClient{
		doFunc: func(_ *http.Request) (*http.Response, error) {
			return nil, errors.New("timeout connecting")
		},
	}

	input := "add http://example.com/bad\nimg\nexit\n"
	var out bytes.Buffer

	r := repl.New(strings.NewReader(input), &out, repl.WithHTTPClient(mockClient))
	err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "Scanning http://example.com/bad => error:") || !strings.Contains(out.String(), "timeout connecting") {
		t.Errorf("expected error message in output, got %q", out.String())
	}
	if r.URLList().Len() != 0 {
		t.Errorf("expected url list to be emptied after img extraction failure, got %d items", r.URLList().Len())
	}
}

func TestREPLImgHTTPError(t *testing.T) {
	mockClient := &mockHTTPClient{
		doFunc: func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("404 not found")),
			}, nil
		},
	}

	input := "add http://example.com/404\nimg\nexit\n"
	var out bytes.Buffer

	r := repl.New(strings.NewReader(input), &out, repl.WithHTTPClient(mockClient))
	err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "Scanning http://example.com/404 => 404") {
		t.Errorf("expected 404 scanning message in output, got %q", out.String())
	}
	if r.URLList().Len() != 0 {
		t.Errorf("expected url list to be emptied after 404, got %d items", r.URLList().Len())
	}
}
