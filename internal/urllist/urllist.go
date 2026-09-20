// Package urllist manages an in-memory list of URLs for URLoad2.
package urllist

import (
	"errors"
	"fmt"
	"net/url"
	"sync"
)

var (
	// ErrInvalidURL is returned when an added URL is malformed or lacks scheme/host.
	ErrInvalidURL = errors.New("invalid URL: must be an absolute http or https URL")
	// ErrInvalidCount is returned when an invalid count is provided to Head or Tail.
	ErrInvalidCount = errors.New("count must be a non-negative integer")
)

// List maintains an ordered collection of URLs.
type List struct {
	mu   sync.RWMutex
	urls []string
}

// New creates a new empty List.
func New() *List {
	return &List{
		urls: make([]string, 0),
	}
}

// Add appends a validated URL to the end of the list.
func (l *List) Add(rawURL string) error {
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return fmt.Errorf("%w: %q", ErrInvalidURL, rawURL)
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.urls = append(l.urls, rawURL)
	return nil
}

// Get returns a copy of the current list of URLs.
func (l *List) Get() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	copied := make([]string, len(l.urls))
	copy(copied, l.urls)
	return copied
}

// Len returns the number of URLs currently in the list.
func (l *List) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.urls)
}

// Clear removes all URLs from the list.
func (l *List) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.urls = make([]string, 0)
}

// Head keeps the first n URLs in the list, discarding the rest.
// If n is greater than or equal to the list length, all URLs are kept.
func (l *List) Head(n int) error {
	if n < 0 {
		return ErrInvalidCount
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if n < len(l.urls) {
		l.urls = l.urls[:n]
	}
	return nil
}

// Tail keeps the last n URLs in the list, discarding the rest.
// If n is greater than or equal to the list length, all URLs are kept.
func (l *List) Tail(n int) error {
	if n < 0 {
		return ErrInvalidCount
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if n < len(l.urls) {
		l.urls = l.urls[len(l.urls)-n:]
	}
	return nil
}
