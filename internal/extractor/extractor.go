// Package extractor extracts and normalizes URLs from HTML pages.
package extractor

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// HTTPGetter defines the HTTP request execution interface.
type HTTPGetter interface {
	Do(req *http.Request) (*http.Response, error)
}

// ResolveURL resolves a raw target URL against a base URL and normalizes it to an absolute HTTP/HTTPS URL.
// It strips URL fragments and filters out non-HTTP schemes (such as mailto:, javascript:).
func ResolveURL(baseRaw, targetRaw string) (string, bool) {
	trimmedTarget := strings.TrimSpace(targetRaw)
	if trimmedTarget == "" {
		return "", false
	}

	base, err := url.Parse(baseRaw)
	if err != nil {
		return "", false
	}

	target, err := url.Parse(trimmedTarget)
	if err != nil {
		return "", false
	}

	resolved := base.ResolveReference(target)
	resolved.Fragment = "" // Strip fragment identifiers

	if (resolved.Scheme != "http" && resolved.Scheme != "https") || resolved.Host == "" {
		return "", false
	}

	return resolved.String(), true
}

// ExtractTargets fetches the HTML at pageURL and extracts all attribute values matching targetTag and targetAttr.
// It returns the extracted targets, the HTTP status code, and any error encountered.
func ExtractTargets(ctx context.Context, client HTTPGetter, pageURL, targetTag, targetAttr string) ([]string, int, error) {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request for %s: %w", pageURL, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch %s: %w", pageURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp.StatusCode, fmt.Errorf("HTTP status %d %s", resp.StatusCode, resp.Status)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to parse HTML from %s: %w", pageURL, err)
	}

	var results []string
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && strings.EqualFold(n.Data, targetTag) {
			for _, attr := range n.Attr {
				if strings.EqualFold(attr.Key, targetAttr) {
					if absURL, ok := ResolveURL(pageURL, attr.Val); ok {
						results = append(results, absURL)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return results, resp.StatusCode, nil
}

// ExtractHrefs extracts all anchor href targets from pageURL normalized as absolute URLs.
func ExtractHrefs(ctx context.Context, client HTTPGetter, pageURL string) ([]string, int, error) {
	return ExtractTargets(ctx, client, pageURL, "a", "href")
}

// ExtractImgs extracts all image src targets from pageURL normalized as absolute URLs.
func ExtractImgs(ctx context.Context, client HTTPGetter, pageURL string) ([]string, int, error) {
	return ExtractTargets(ctx, client, pageURL, "img", "src")
}
