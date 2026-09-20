package urllist_test

import (
	"fmt"
	"testing"

	"pgregory.net/rapid"

	"github.com/JohnStrunk/URLoad2/internal/urllist"
)

func genValidURL() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		scheme := rapid.SampledFrom([]string{"http", "https"}).Draw(t, "scheme")
		host := rapid.StringMatching(`[a-z0-9]{1,10}\.[a-z]{2,4}`).Draw(t, "host")
		path := rapid.StringMatching(`(/[a-z0-9]{1,8})*`).Draw(t, "path")
		return fmt.Sprintf("%s://%s%s", scheme, host, path)
	})
}

// EARS: When the user enters the add command with a valid URL, the REPL shall append the URL to the end of the URL list.
func TestProperty_EARS_AddPreservesOrderAndLength(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		urls := rapid.SliceOfN(genValidURL(), 0, 30).Draw(t, "urls")
		l := urllist.New()

		for i, u := range urls {
			err := l.Add(u)
			if err != nil {
				t.Fatalf("unexpected error adding valid URL %q: %v", u, err)
			}
			if l.Len() != i+1 {
				t.Fatalf("expected len %d, got %d", i+1, l.Len())
			}
		}

		got := l.Get()
		if len(got) != len(urls) {
			t.Fatalf("expected len %d, got %d", len(urls), len(got))
		}
		for i := range urls {
			if got[i] != urls[i] {
				t.Fatalf("at index %d: expected %q, got %q", i, urls[i], got[i])
			}
		}
	})
}

// EARS: When the user enters the clear command, the REPL shall remove all URLs from the list.
func TestProperty_EARS_ClearEmptiesList(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		urls := rapid.SliceOfN(genValidURL(), 0, 30).Draw(t, "urls")
		l := urllist.New()
		for _, u := range urls {
			_ = l.Add(u)
		}

		l.Clear()

		if l.Len() != 0 {
			t.Fatalf("expected len 0 after clear, got %d", l.Len())
		}
		if len(l.Get()) != 0 {
			t.Fatalf("expected Get() to be empty after clear, got %v", l.Get())
		}
	})
}

// EARS: When the user enters the head command with count n, the REPL shall retain the first n URLs in the list and discard the rest.
func TestProperty_EARS_HeadInvariants(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		urls := rapid.SliceOfN(genValidURL(), 0, 30).Draw(t, "urls")
		l := urllist.New()
		for _, u := range urls {
			_ = l.Add(u)
		}

		n := rapid.IntRange(0, 50).Draw(t, "n")
		err := l.Head(n)
		if err != nil {
			t.Fatalf("unexpected error for non-negative n=%d: %v", n, err)
		}

		expectedLen := n
		if expectedLen > len(urls) {
			expectedLen = len(urls)
		}

		if l.Len() != expectedLen {
			t.Fatalf("expected len %d, got %d", expectedLen, l.Len())
		}

		got := l.Get()
		for i := 0; i < expectedLen; i++ {
			if got[i] != urls[i] {
				t.Fatalf("at index %d: expected %q, got %q", i, urls[i], got[i])
			}
		}
	})
}

// EARS: When the user enters the tail command with count n, the REPL shall retain the last n URLs in the list and discard the rest.
func TestProperty_EARS_TailInvariants(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		urls := rapid.SliceOfN(genValidURL(), 0, 30).Draw(t, "urls")
		l := urllist.New()
		for _, u := range urls {
			_ = l.Add(u)
		}

		n := rapid.IntRange(0, 50).Draw(t, "n")
		err := l.Tail(n)
		if err != nil {
			t.Fatalf("unexpected error for non-negative n=%d: %v", n, err)
		}

		expectedLen := n
		if expectedLen > len(urls) {
			expectedLen = len(urls)
		}

		if l.Len() != expectedLen {
			t.Fatalf("expected len %d, got %d", expectedLen, l.Len())
		}

		got := l.Get()
		startIdx := len(urls) - expectedLen
		for i := 0; i < expectedLen; i++ {
			if got[i] != urls[startIdx+i] {
				t.Fatalf("at index %d: expected %q, got %q", i, urls[startIdx+i], got[i])
			}
		}
	})
}
