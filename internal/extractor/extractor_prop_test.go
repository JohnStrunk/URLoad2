package extractor_test

import (
	"fmt"
	"net/url"
	"strings"
	"testing"

	"pgregory.net/rapid"

	"github.com/JohnStrunk/URLoad2/internal/extractor"
)

func genBaseURL() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		scheme := rapid.SampledFrom([]string{"http", "https"}).Draw(t, "scheme")
		host := rapid.StringMatching(`[a-z0-9]{1,8}\.[a-z]{2,4}`).Draw(t, "host")
		path := rapid.StringMatching(`(/[a-z0-9]{1,6})*`).Draw(t, "path")
		return fmt.Sprintf("%s://%s%s", scheme, host, path)
	})
}

// EARS: When resolving URLs, the extractor shall produce valid absolute HTTP or HTTPS URLs with fragments stripped.
func TestProperty_EARS_ResolveURLInvariants(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		base := genBaseURL().Draw(rt, "base")
		targetRel := rapid.StringMatching(`[a-zA-Z0-9_\-\./]{1,20}`).Draw(rt, "targetRel")
		fragment := rapid.StringMatching(`(#[a-zA-Z0-9_\-]{0,10})?`).Draw(rt, "fragment")

		rawTarget := targetRel + fragment
		resolved, ok := extractor.ResolveURL(base, rawTarget)
		if !ok {
			t.Fatalf("expected valid resolution for base %q and target %q", base, rawTarget)
		}

		// 1. Must parse cleanly
		parsed, err := url.Parse(resolved)
		if err != nil {
			t.Fatalf("resolved URL %q cannot be parsed: %v", resolved, err)
		}

		// 2. Must have scheme http or https
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			t.Fatalf("resolved URL %q has invalid scheme: %q", resolved, parsed.Scheme)
		}

		// 3. Must have non-empty host
		if parsed.Host == "" {
			t.Fatalf("resolved URL %q has empty host", resolved)
		}

		// 4. Must not contain fragment
		if parsed.Fragment != "" || strings.Contains(resolved, "#") {
			t.Fatalf("resolved URL %q still contains fragment", resolved)
		}
	})
}
