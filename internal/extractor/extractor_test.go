package extractor_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JohnStrunk/URLoad2/internal/extractor"
)

func TestResolveURL(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		target   string
		expected string
		ok       bool
	}{
		{
			name:     "relative path",
			base:     "https://example.com/dir/page.html",
			target:   "other.html",
			expected: "https://example.com/dir/other.html",
			ok:       true,
		},
		{
			name:     "root-relative path",
			base:     "https://example.com/dir/page.html",
			target:   "/root.html",
			expected: "https://example.com/root.html",
			ok:       true,
		},
		{
			name:     "absolute URL",
			base:     "https://example.com/",
			target:   "http://other.org/test",
			expected: "http://other.org/test",
			ok:       true,
		},
		{
			name:     "strip fragment",
			base:     "https://example.com/page",
			target:   "/about#section1",
			expected: "https://example.com/about",
			ok:       true,
		},
		{
			name:     "empty target",
			base:     "https://example.com/",
			target:   "",
			expected: "",
			ok:       false,
		},
		{
			name:     "whitespace target",
			base:     "https://example.com/",
			target:   "   ",
			expected: "",
			ok:       false,
		},
		{
			name:     "mailto scheme ignored",
			base:     "https://example.com/",
			target:   "mailto:info@example.com",
			expected: "",
			ok:       false,
		},
		{
			name:     "javascript scheme ignored",
			base:     "https://example.com/",
			target:   "javascript:alert(1)",
			expected: "",
			ok:       false,
		},
		{
			name:     "invalid base URL",
			base:     "://bad-base",
			target:   "/path",
			expected: "",
			ok:       false,
		},
		{
			name:     "invalid target URL",
			base:     "https://example.com",
			target:   "://bad-target",
			expected: "",
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := extractor.ResolveURL(tt.base, tt.target)
			if ok != tt.ok {
				t.Fatalf("expected ok=%v, got %v", tt.ok, ok)
			}
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestExtractHrefs(t *testing.T) {
	htmlContent := `
<!DOCTYPE html>
<html>
<body>
	<a href="/page1.html">Page 1</a>
	<a href="page2.html#anchor">Page 2 with anchor</a>
	<a href="https://external.com/api">External</a>
	<a href="mailto:test@example.com">Email</a>
	<a href="javascript:void(0)">JS</a>
	<a name="no-href">No href</a>
	<div>
		<a href="nested/deep.html">Deep</a>
	</div>
</body>
</html>`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, htmlContent)
	}))
	defer ts.Close()

	urls, err := extractor.ExtractHrefs(context.Background(), nil, ts.URL+"/sub/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		ts.URL + "/page1.html",
		ts.URL + "/sub/page2.html",
		"https://external.com/api",
		ts.URL + "/sub/nested/deep.html",
	}

	if len(urls) != len(expected) {
		t.Fatalf("expected %d urls, got %d: %v", len(expected), len(urls), urls)
	}

	for i := range expected {
		if urls[i] != expected[i] {
			t.Errorf("at index %d: expected %q, got %q", i, expected[i], urls[i])
		}
	}
}

func TestExtractImgs(t *testing.T) {
	htmlContent := `
<!DOCTYPE html>
<html>
<body>
	<img src="/img/logo.png" alt="Logo">
	<img src="banner.jpg">
	<img src="https://cdn.example.com/pic.webp">
	<img alt="No src">
	<img src="data:image/png;base64,xxxx">
	<div>
		<p><img src="../common/icon.svg"></p>
	</div>
</body>
</html>`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, htmlContent)
	}))
	defer ts.Close()

	urls, err := extractor.ExtractImgs(context.Background(), ts.Client(), ts.URL+"/dir/page.html")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		ts.URL + "/img/logo.png",
		ts.URL + "/dir/banner.jpg",
		"https://cdn.example.com/pic.webp",
		ts.URL + "/common/icon.svg",
	}

	if len(urls) != len(expected) {
		t.Fatalf("expected %d urls, got %d: %v", len(expected), len(urls), urls)
	}

	for i := range expected {
		if urls[i] != expected[i] {
			t.Errorf("at index %d: expected %q, got %q", i, expected[i], urls[i])
		}
	}
}

func TestExtractHTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Not found", http.StatusNotFound)
	}))
	defer ts.Close()

	_, err := extractor.ExtractHrefs(context.Background(), ts.Client(), ts.URL+"/missing")
	if err == nil {
		t.Fatal("expected error on 404, got nil")
	}
}

type errClient struct {
	err error
}

func (e *errClient) Do(_ *http.Request) (*http.Response, error) {
	return nil, e.err
}

func TestExtractNetworkError(t *testing.T) {
	client := &errClient{err: errors.New("connection failed")}
	_, err := extractor.ExtractHrefs(context.Background(), client, "http://example.com")
	if err == nil {
		t.Fatal("expected error on network failure, got nil")
	}
}

func TestExtractInvalidURL(t *testing.T) {
	_, err := extractor.ExtractHrefs(context.Background(), nil, "://bad-url")
	if err == nil {
		t.Fatal("expected error on invalid URL, got nil")
	}
}
