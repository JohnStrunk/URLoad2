package downloader_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"pgregory.net/rapid"

	"github.com/JohnStrunk/URLoad2/internal/downloader"
)

// EARS: When the REPL starts, the REPL shall determine the download target directory as the lowest
// four-digit integer string starting at 0000 that is one greater than the highest existing four-digit
// directory in the current working directory.
func TestProperty_EARS_DetermineTargetDir(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		tempDir, err := os.MkdirTemp("", "rapid-downloader-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Generate random list of numbers between 0 and 5000
		nums := rapid.SliceOfN(rapid.IntRange(0, 5000), 0, 20).Draw(rt, "existingDirs")

		maxNum := -1
		for _, n := range nums {
			dirName := fmt.Sprintf("%04d", n)
			_ = os.Mkdir(filepath.Join(tempDir, dirName), 0755)
			if n > maxNum {
				maxNum = n
			}
		}

		// Also generate some random non-numeric directories to ensure they don't interfere
		nonNumDirs := rapid.SliceOfN(rapid.StringMatching(`[a-zA-Z]{1,8}`), 0, 5).Draw(rt, "nonNumericDirs")
		for _, d := range nonNumDirs {
			_ = os.Mkdir(filepath.Join(tempDir, d), 0755)
		}

		target, err := downloader.DetermineTargetDir(tempDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedNum := maxNum + 1
		expectedStr := fmt.Sprintf("%04d", expectedNum)

		if target != expectedStr {
			t.Fatalf("expected target %q, got %q (max was %d)", expectedStr, target, maxNum)
		}

		// Verify parsed integer
		parsed, err := strconv.Atoi(target)
		if err != nil {
			t.Fatalf("target %q is not a valid integer: %v", target, err)
		}
		if parsed != expectedNum {
			t.Fatalf("expected integer %d, got %d", expectedNum, parsed)
		}
	})
}
