package urllist_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/JohnStrunk/URLoad2/internal/urllist"
)

func TestNewListIsEmpty(t *testing.T) {
	l := urllist.New()
	if l.Len() != 0 {
		t.Fatalf("expected empty list, got len %d", l.Len())
	}
	if len(l.Get()) != 0 {
		t.Fatalf("expected empty slice from Get(), got %v", l.Get())
	}
}

func TestAddValidURLs(t *testing.T) {
	l := urllist.New()

	validURLs := []string{
		"http://example.com",
		"https://example.com/path",
		"https://sub.domain.org/path?query=val#frag",
	}

	for _, u := range validURLs {
		if err := l.Add(u); err != nil {
			t.Fatalf("expected nil error adding %q, got %v", u, err)
		}
	}

	if l.Len() != len(validURLs) {
		t.Fatalf("expected len %d, got %d", len(validURLs), l.Len())
	}

	got := l.Get()
	for i, u := range validURLs {
		if got[i] != u {
			t.Errorf("at index %d: expected %q, got %q", i, u, got[i])
		}
	}
}

func TestAddInvalidURLs(t *testing.T) {
	l := urllist.New()

	invalidURLs := []string{
		"",
		"not-a-url",
		"ftp://example.com",
		"/just/a/path",
		"http://",
		"https://",
	}

	for _, u := range invalidURLs {
		err := l.Add(u)
		if err == nil {
			t.Errorf("expected error adding invalid URL %q, got nil", u)
		}
		if !errors.Is(err, urllist.ErrInvalidURL) {
			t.Errorf("expected ErrInvalidURL, got %v", err)
		}
	}

	if l.Len() != 0 {
		t.Fatalf("expected list to remain empty, got len %d", l.Len())
	}
}

func TestGetReturnsCopy(t *testing.T) {
	l := urllist.New()
	_ = l.Add("https://example.com/1")

	items := l.Get()
	items[0] = "mutated"

	got := l.Get()
	if got[0] != "https://example.com/1" {
		t.Fatalf("Get did not return a safe copy; got %q", got[0])
	}
}

func TestClear(t *testing.T) {
	l := urllist.New()
	_ = l.Add("https://example.com/1")
	_ = l.Add("https://example.com/2")

	if l.Len() != 2 {
		t.Fatalf("expected len 2, got %d", l.Len())
	}

	l.Clear()

	if l.Len() != 0 {
		t.Fatalf("expected len 0 after clear, got %d", l.Len())
	}
	if len(l.Get()) != 0 {
		t.Fatalf("expected empty slice after clear, got %v", l.Get())
	}
}

func TestHead(t *testing.T) {
	tests := []struct {
		name        string
		initial     []string
		n           int
		expected    []string
		expectError bool
	}{
		{
			name:     "head keeps first n elements",
			initial:  []string{"http://a.com", "http://b.com", "http://c.com"},
			n:        2,
			expected: []string{"http://a.com", "http://b.com"},
		},
		{
			name:     "head 0 empties list",
			initial:  []string{"http://a.com", "http://b.com"},
			n:        0,
			expected: []string{},
		},
		{
			name:     "head with n greater than len keeps all",
			initial:  []string{"http://a.com", "http://b.com"},
			n:        5,
			expected: []string{"http://a.com", "http://b.com"},
		},
		{
			name:     "head with n equal to len keeps all",
			initial:  []string{"http://a.com", "http://b.com"},
			n:        2,
			expected: []string{"http://a.com", "http://b.com"},
		},
		{
			name:        "head with negative n returns error",
			initial:     []string{"http://a.com"},
			n:           -1,
			expected:    []string{"http://a.com"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := urllist.New()
			for _, u := range tt.initial {
				_ = l.Add(u)
			}

			err := l.Head(tt.n)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !errors.Is(err, urllist.ErrInvalidCount) {
					t.Fatalf("expected ErrInvalidCount, got %v", err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := l.Get()
			if len(got) != len(tt.expected) {
				t.Fatalf("expected len %d, got %d: %v", len(tt.expected), len(got), got)
			}
			for i := range tt.expected {
				if got[i] != tt.expected[i] {
					t.Errorf("at %d: expected %q, got %q", i, tt.expected[i], got[i])
				}
			}
		})
	}
}

func TestTail(t *testing.T) {
	tests := []struct {
		name        string
		initial     []string
		n           int
		expected    []string
		expectError bool
	}{
		{
			name:     "tail keeps last n elements",
			initial:  []string{"http://a.com", "http://b.com", "http://c.com"},
			n:        2,
			expected: []string{"http://b.com", "http://c.com"},
		},
		{
			name:     "tail 0 empties list",
			initial:  []string{"http://a.com", "http://b.com"},
			n:        0,
			expected: []string{},
		},
		{
			name:     "tail with n greater than len keeps all",
			initial:  []string{"http://a.com", "http://b.com"},
			n:        5,
			expected: []string{"http://a.com", "http://b.com"},
		},
		{
			name:     "tail with n equal to len keeps all",
			initial:  []string{"http://a.com", "http://b.com"},
			n:        2,
			expected: []string{"http://a.com", "http://b.com"},
		},
		{
			name:        "tail with negative n returns error",
			initial:     []string{"http://a.com"},
			n:           -1,
			expected:    []string{"http://a.com"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := urllist.New()
			for _, u := range tt.initial {
				_ = l.Add(u)
			}

			err := l.Tail(tt.n)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !errors.Is(err, urllist.ErrInvalidCount) {
					t.Fatalf("expected ErrInvalidCount, got %v", err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := l.Get()
			if len(got) != len(tt.expected) {
				t.Fatalf("expected len %d, got %d: %v", len(tt.expected), len(got), got)
			}
			for i := range tt.expected {
				if got[i] != tt.expected[i] {
					t.Errorf("at %d: expected %q, got %q", i, tt.expected[i], got[i])
				}
			}
		})
	}
}

func TestConcurrentAccess(t *testing.T) {
	l := urllist.New()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_ = l.Add(fmt.Sprintf("https://example.com/%d", idx))
			_ = l.Len()
			_ = l.Get()
		}(i)
	}

	wg.Wait()

	if l.Len() != 50 {
		t.Fatalf("expected len 50 after concurrent adds, got %d", l.Len())
	}
}
