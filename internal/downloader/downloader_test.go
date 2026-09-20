package downloader_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JohnStrunk/URLoad2/internal/downloader"
)

func TestDetermineTargetDir(t *testing.T) {
	t.Run("empty directory starts at 0000", func(t *testing.T) {
		tempDir := t.TempDir()
		target, err := downloader.DetermineTargetDir(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if target != "0000" {
			t.Fatalf("expected 0000, got %q", target)
		}
	})

	t.Run("non-existent directory defaults to 0000", func(t *testing.T) {
		nonExistent := filepath.Join(t.TempDir(), "not-exists")
		target, err := downloader.DetermineTargetDir(nonExistent)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if target != "0000" {
			t.Fatalf("expected 0000, got %q", target)
		}
	})

	t.Run("increments from highest existing four-digit directory", func(t *testing.T) {
		tempDir := t.TempDir()
		dirs := []string{"0000", "0003", "0007", "not-a-num", "123", "12345"}
		for _, d := range dirs {
			if err := os.Mkdir(filepath.Join(tempDir, d), 0755); err != nil {
				t.Fatalf("failed to create dir %s: %v", d, err)
			}
		}

		// Also create a regular file named 0009; it should be ignored because it is not a directory
		filePath := filepath.Join(tempDir, "0009")
		if err := os.WriteFile(filePath, []byte("data"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}

		target, err := downloader.DetermineTargetDir(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Highest 4-digit directory is 0007, so next is 0008
		if target != "0008" {
			t.Fatalf("expected 0008, got %q", target)
		}
	})
}

func TestDeferredCreation(t *testing.T) {
	tempDir := t.TempDir()

	dl, err := downloader.New(downloader.WithBaseDir(tempDir))
	if err != nil {
		t.Fatalf("failed to create downloader: %v", err)
	}

	if dl.TargetName() != "0000" {
		t.Errorf("expected target name 0000, got %s", dl.TargetName())
	}
	if dl.IsCreated() {
		t.Errorf("expected IsCreated to be false initially")
	}

	// Verify target dir does not exist on disk
	if _, err := os.Stat(dl.TargetPath()); !os.IsNotExist(err) {
		t.Fatalf("expected target directory %s to not exist yet", dl.TargetPath())
	}

	// Calling DownloadAll with no URLs should not create directory
	var buf bytes.Buffer
	err = dl.DownloadAll(context.Background(), nil, &buf)
	if err != nil {
		t.Fatalf("unexpected error downloading nil list: %v", err)
	}
	if dl.IsCreated() {
		t.Errorf("expected IsCreated to remain false after empty download")
	}
	if _, err := os.Stat(dl.TargetPath()); !os.IsNotExist(err) {
		t.Fatalf("expected target directory to still not exist after empty download")
	}

	// Setup mock server to serve 1 URL
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "Hello, URLoad2!")
	}))
	defer ts.Close()

	buf.Reset()
	err = dl.DownloadAll(context.Background(), []string{ts.URL + "/page.html"}, &buf)
	if err != nil {
		t.Fatalf("failed to download: %v", err)
	}

	if !dl.IsCreated() {
		t.Errorf("expected IsCreated to be true after successful download")
	}
	info, err := os.Stat(dl.TargetPath())
	if err != nil || !info.IsDir() {
		t.Fatalf("expected target directory %s to exist as a directory, err: %v", dl.TargetPath(), err)
	}

	savedFile := filepath.Join(dl.TargetPath(), "page.html")
	content, err := os.ReadFile(savedFile)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if !strings.Contains(string(content), "Hello, URLoad2!") {
		t.Fatalf("unexpected file content: %q", string(content))
	}
}

func TestDownloadUniqueFilenames(t *testing.T) {
	tempDir := t.TempDir()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "content from %s", r.URL.Path)
	}))
	defer ts.Close()

	dl, err := downloader.New(downloader.WithBaseDir(tempDir))
	if err != nil {
		t.Fatalf("failed to create downloader: %v", err)
	}

	urls := []string{
		ts.URL + "/docs/guide.html",
		ts.URL + "/blog/guide.html",
		ts.URL + "/api/guide.html",
		ts.URL + "/", // Should become index.html
	}

	var buf bytes.Buffer
	err = dl.DownloadAll(context.Background(), urls, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	files, err := os.ReadDir(dl.TargetPath())
	if err != nil {
		t.Fatalf("failed to read target directory: %v", err)
	}

	expectedFiles := map[string]bool{
		"guide.html":   false,
		"guide_1.html": false,
		"guide_2.html": false,
		"index.html":   false,
	}

	for _, f := range files {
		expectedFiles[f.Name()] = true
	}

	for name, found := range expectedFiles {
		if !found {
			t.Errorf("expected file %s to be created", name)
		}
	}
}

func TestDownloadFailureContinues(t *testing.T) {
	tempDir := t.TempDir()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/fail" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		fmt.Fprintln(w, "success")
	}))
	defer ts.Close()

	dl, err := downloader.New(downloader.WithBaseDir(tempDir))
	if err != nil {
		t.Fatalf("failed to create downloader: %v", err)
	}

	urls := []string{
		ts.URL + "/fail",
		ts.URL + "/good.html",
	}

	var buf bytes.Buffer
	err = dl.DownloadAll(context.Background(), urls, &buf)
	if err != nil {
		t.Fatalf("DownloadAll returned unexpected error: %v", err)
	}

	output := buf.String()
	expected404 := fmt.Sprintf("Downloading %s/fail => 404", ts.URL)
	if !strings.Contains(output, expected404) {
		t.Errorf("expected 404 message in output %q, got: %s", expected404, output)
	}
	expectedGood := fmt.Sprintf("Downloading %s/good.html => 200 [good.html]", ts.URL)
	if !strings.Contains(output, expectedGood) {
		t.Errorf("expected success message in output %q, got: %s", expectedGood, output)
	}

	// Good file should exist
	if _, err := os.Stat(filepath.Join(dl.TargetPath(), "good.html")); err != nil {
		t.Fatalf("expected good.html to exist, err: %v", err)
	}
}

type failingHTTPClient struct {
	err error
}

func (f *failingHTTPClient) Do(_ *http.Request) (*http.Response, error) {
	return nil, f.err
}

func TestDownloadNetworkError(t *testing.T) {
	tempDir := t.TempDir()
	dl, err := downloader.New(
		downloader.WithBaseDir(tempDir),
		downloader.WithHTTPClient(&failingHTTPClient{err: errors.New("connection refused")}),
	)
	if err != nil {
		t.Fatalf("failed to create downloader: %v", err)
	}

	var buf bytes.Buffer
	err = dl.DownloadAll(context.Background(), []string{"http://broken.test/test"}, &buf)
	if err != nil {
		t.Fatalf("unexpected top-level error: %v", err)
	}

	if !strings.Contains(buf.String(), "Downloading http://broken.test/test => error: connection refused") {
		t.Fatalf("expected error output, got %q", buf.String())
	}
}

func TestDownloadContextCanceled(t *testing.T) {
	tempDir := t.TempDir()
	dl, err := downloader.New(downloader.WithBaseDir(tempDir))
	if err != nil {
		t.Fatalf("failed to create downloader: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var buf bytes.Buffer
	err = dl.DownloadAll(ctx, []string{"http://example.com/test"}, &buf)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestNewDefault(t *testing.T) {
	dl, err := downloader.New()
	if err != nil {
		t.Fatalf("unexpected error creating default downloader: %v", err)
	}
	if dl.TargetName() == "" {
		t.Errorf("expected non-empty target name")
	}
}

func TestDetermineTargetDirReadDirError(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "regular_file")
	if err := os.WriteFile(tempFile, []byte("data"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	_, err := downloader.DetermineTargetDir(tempFile)
	if err == nil {
		t.Fatalf("expected error reading regular file as directory, got nil")
	}
}

type errWriter struct {
	err error
}

func (w *errWriter) Write(_ []byte) (int, error) {
	return 0, w.err
}

func TestDownloadAllWriteErrors(t *testing.T) {
	tempDir := t.TempDir()
	dl, err := downloader.New(downloader.WithBaseDir(tempDir))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedErr := errors.New("cannot write")
	err = dl.DownloadAll(context.Background(), nil, &errWriter{err: expectedErr})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected write error on empty download, got %v", err)
	}
}

func TestDownloadAllMkdirError(t *testing.T) {
	tempDir := t.TempDir()
	// Create a file where the target dir should be
	targetFile := filepath.Join(tempDir, "0000")
	if err := os.WriteFile(targetFile, []byte("blocker"), 0644); err != nil {
		t.Fatalf("failed to create blocking file: %v", err)
	}

	dl, err := downloader.New(downloader.WithBaseDir(tempDir), downloader.WithTargetName("0000"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	err = dl.DownloadAll(context.Background(), []string{"http://example.com/test"}, &buf)
	if err == nil || !strings.Contains(err.Error(), "failed to create target directory") {
		t.Fatalf("expected failed to create target directory error, got %v", err)
	}
}

func TestDownloadOneInvalidURL(t *testing.T) {
	tempDir := t.TempDir()
	dl, err := downloader.New(downloader.WithBaseDir(tempDir))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	err = dl.DownloadAll(context.Background(), []string{"://invalid-url"}, &buf)
	if err != nil {
		t.Fatalf("expected nil top-level error on per-URL failure, got %v", err)
	}
	if !strings.Contains(buf.String(), "Downloading ://invalid-url => error:") {
		t.Fatalf("expected error message in output, got: %s", buf.String())
	}
}

func TestDownloadOneWriteErrors(t *testing.T) {
	tempDir := t.TempDir()
	dl, err := downloader.New(downloader.WithBaseDir(tempDir))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedErr := errors.New("cannot write")
	err = dl.DownloadAll(context.Background(), []string{"://invalid-url"}, &errWriter{err: expectedErr})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected write error on invalid url, got %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/404" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		fmt.Fprintln(w, "ok")
	}))
	defer ts.Close()

	err = dl.DownloadAll(context.Background(), []string{ts.URL + "/404"}, &errWriter{err: expectedErr})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected write error on 404, got %v", err)
	}

	err = dl.DownloadAll(context.Background(), []string{ts.URL + "/200"}, &errWriter{err: expectedErr})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected write error on 200, got %v", err)
	}
}

func TestWithTargetName(t *testing.T) {
	tempDir := t.TempDir()
	dl, err := downloader.New(
		downloader.WithBaseDir(tempDir),
		downloader.WithTargetName("4321"),
	)
	if err != nil {
		t.Fatalf("failed to create downloader: %v", err)
	}

	if dl.TargetName() != "4321" {
		t.Fatalf("expected target name 4321, got %s", dl.TargetName())
	}
}
