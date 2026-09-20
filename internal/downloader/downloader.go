// Package downloader handles determining download target directories and fetching URLs.
package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var fourDigitRegex = regexp.MustCompile(`^[0-9]{4}$`)

// HTTPGetter defines the HTTP request execution interface.
type HTTPGetter interface {
	Do(req *http.Request) (*http.Response, error)
}

// Downloader manages downloading files to a session-specific target directory.
type Downloader struct {
	mu         sync.Mutex
	baseDir    string
	targetName string
	targetPath string
	created    bool
	client     HTTPGetter
}

// Option configures a Downloader.
type Option func(*Downloader)

// WithBaseDir sets the base directory (defaults to cwd).
func WithBaseDir(dir string) Option {
	return func(d *Downloader) {
		d.baseDir = dir
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client HTTPGetter) Option {
	return func(d *Downloader) {
		d.client = client
	}
}

// WithTargetName overrides the calculated target name (primarily for testing).
func WithTargetName(name string) Option {
	return func(d *Downloader) {
		d.targetName = name
	}
}

// DetermineTargetDir inspects baseDir for existing 4-digit directories (<nnnn>)
// and returns the next 4-digit directory name (0000 or highest+1).
func DetermineTargetDir(baseDir string) (string, error) {
	info, err := os.Stat(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "0000", nil
		}
		return "", fmt.Errorf("failed to stat directory: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", baseDir)
	}

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	maxNum := -1
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if fourDigitRegex.MatchString(name) {
			num, err := strconv.Atoi(name)
			if err == nil && num > maxNum {
				maxNum = num
			}
		}
	}

	nextNum := maxNum + 1
	return fmt.Sprintf("%04d", nextNum), nil
}

// New creates a Downloader, calculating the target directory name without creating it.
func New(opts ...Option) (*Downloader, error) {
	d := &Downloader{
		client: &http.Client{Timeout: 30 * time.Second},
	}

	for _, opt := range opts {
		opt(d)
	}

	if d.baseDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
		d.baseDir = cwd
	}

	if d.targetName == "" {
		targetName, err := DetermineTargetDir(d.baseDir)
		if err != nil {
			return nil, err
		}
		d.targetName = targetName
	}

	d.targetPath = filepath.Join(d.baseDir, d.targetName)
	return d, nil
}

// TargetName returns the 4-digit target directory name.
func (d *Downloader) TargetName() string {
	return d.targetName
}

// TargetPath returns the absolute path to the target directory.
func (d *Downloader) TargetPath() string {
	return d.targetPath
}

// IsCreated returns whether the target directory has been created yet.
func (d *Downloader) IsCreated() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.created
}

// sanitizeFilename generates a valid local filename from a URL.
func sanitizeFilename(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "download.html"
	}

	base := path.Base(parsed.Path)
	if base == "" || base == "/" || base == "." {
		base = "index.html"
	}

	// Remove invalid filesystem characters
	invalidChars := `/\:*?"<>|`
	var sb strings.Builder
	for _, r := range base {
		if strings.ContainsRune(invalidChars, r) {
			sb.WriteRune('_')
		} else {
			sb.WriteRune(r)
		}
	}

	cleaned := sb.String()
	if cleaned == "" || cleaned == "." {
		cleaned = "index.html"
	}
	return cleaned
}

// uniqueFilename ensures the filename is not already used in this directory.
func uniqueFilename(dir, baseName string, usedNames map[string]bool) string {
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)

	candidate := baseName
	counter := 1
	for {
		_, exists := usedNames[candidate]
		if !exists {
			filePath := filepath.Join(dir, candidate)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				usedNames[candidate] = true
				return candidate
			}
		}
		candidate = fmt.Sprintf("%s_%d%s", nameWithoutExt, counter, ext)
		counter++
	}
}

// DownloadAll downloads all provided URLs to the target directory, creating it if needed.
func (d *Downloader) DownloadAll(ctx context.Context, urls []string, out io.Writer) error {
	if len(urls) == 0 {
		if _, err := fmt.Fprintln(out, "No URLs to download."); err != nil {
			return err
		}
		return nil
	}

	d.mu.Lock()
	if !d.created {
		if err := os.MkdirAll(d.targetPath, 0755); err != nil {
			d.mu.Unlock()
			return fmt.Errorf("failed to create target directory %s: %w", d.targetPath, err)
		}
		d.created = true
	}
	d.mu.Unlock()

	usedNames := make(map[string]bool)

	for _, rawURL := range urls {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := d.downloadOne(ctx, rawURL, usedNames, out)
		if err != nil {
			if _, writeErr := fmt.Fprintf(out, "error downloading %s: %v\n", rawURL, err); writeErr != nil {
				return writeErr
			}
		}
	}

	return nil
}

func (d *Downloader) downloadOne(ctx context.Context, rawURL string, usedNames map[string]bool, out io.Writer) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP status %d %s", resp.StatusCode, resp.Status)
	}

	baseName := sanitizeFilename(rawURL)
	fileName := uniqueFilename(d.targetPath, baseName, usedNames)
	destPath := filepath.Join(d.targetPath, fileName)

	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", fileName, err)
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("failed to write content to %s: %w", fileName, err)
	}

	if _, err := fmt.Fprintf(out, "Downloaded %s -> %s\n", rawURL, filepath.Join(d.targetName, fileName)); err != nil {
		return err
	}

	return nil
}
