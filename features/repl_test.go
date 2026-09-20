package features_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"github.com/JohnStrunk/URLoad2/internal/downloader"
	"github.com/JohnStrunk/URLoad2/internal/repl"
	"github.com/JohnStrunk/URLoad2/internal/urllist"
)

type replFeature struct {
	in             *bytes.Buffer
	out            *bytes.Buffer
	err            error
	completions    []string
	workingDir     string
	server         *httptest.Server
	serverRoutes   map[string]string
	serverStatuses map[string]int
	urlList        *urllist.List
	downloader     *downloader.Downloader
}

func (r *replFeature) reset() {
	if r.server != nil {
		r.server.Close()
		r.server = nil
	}
	r.serverRoutes = nil
	r.serverStatuses = nil
	if r.workingDir != "" {
		_ = os.RemoveAll(r.workingDir)
		r.workingDir = ""
	}
	r.in = &bytes.Buffer{}
	r.out = &bytes.Buffer{}
	r.err = nil
	r.completions = nil
	r.urlList = urllist.New()
	r.downloader = nil
}

func (r *replFeature) theREPLApplicationIsInitialized() error {
	r.in = &bytes.Buffer{}
	r.out = &bytes.Buffer{}
	r.err = nil
	r.completions = nil
	if r.workingDir == "" {
		tmp, err := os.MkdirTemp("", "godog-repl-*")
		if err != nil {
			return err
		}
		r.workingDir = tmp
	}
	dl, err := downloader.New(downloader.WithBaseDir(r.workingDir))
	if err != nil {
		return err
	}
	r.downloader = dl
	r.urlList = urllist.New()
	return nil
}

func (r *replFeature) theREPLIsStartedWithNoInput() error {
	app := repl.New(
		r.in,
		r.out,
		repl.WithURLList(r.urlList),
		repl.WithDownloader(r.downloader),
	)
	r.err = app.Run(context.Background())
	return nil
}

func (r *replFeature) theUserEnters(input string) error {
	if r.server != nil {
		input = strings.ReplaceAll(input, "<server>", r.server.URL)
	}
	r.in.WriteString(input + "\n")
	app := repl.New(
		r.in,
		r.out,
		repl.WithURLList(r.urlList),
		repl.WithDownloader(r.downloader),
	)
	r.err = app.Run(context.Background())
	return nil
}

func (r *replFeature) theUserRequestsCompletionFor(prefix string) error {
	r.completions = repl.Complete(prefix)
	return nil
}

func (r *replFeature) theOutputShouldContain(expected string) error {
	if r.server != nil {
		expected = strings.ReplaceAll(expected, "<server>", r.server.URL)
	}
	actual := r.out.String()
	if !strings.Contains(actual, expected) {
		return fmt.Errorf("expected output to contain %q, but got %q", expected, actual)
	}
	return nil
}

func (r *replFeature) theOutputShouldNotContain(expected string) error {
	if r.server != nil {
		expected = strings.ReplaceAll(expected, "<server>", r.server.URL)
	}
	actual := r.out.String()
	if strings.Contains(actual, expected) {
		return fmt.Errorf("expected output to not contain %q, but got %q", expected, actual)
	}
	return nil
}

func (r *replFeature) theOutputShouldReportUnknownCommand(cmd string) error {
	expected := fmt.Sprintf("unknown command: %q", cmd)
	return r.theOutputShouldContain(expected)
}

func (r *replFeature) theCompletionCandidatesShouldInclude(expected string) error {
	for _, candidate := range r.completions {
		if candidate == expected {
			return nil
		}
	}
	return fmt.Errorf("expected completion candidates %v to include %q", r.completions, expected)
}

func (r *replFeature) theREPLSessionShouldTerminateSuccessfully() error {
	if r.err != nil {
		return fmt.Errorf("expected clean termination, but got error: %w", r.err)
	}
	return nil
}

func (r *replFeature) aWorkingDirectoryWithDirectory(dirName string) error {
	tmp, err := os.MkdirTemp("", "godog-target-*")
	if err != nil {
		return err
	}
	r.workingDir = tmp
	return os.Mkdir(filepath.Join(r.workingDir, dirName), 0755)
}

func (r *replFeature) theREPLApplicationIsInitializedInThatWorkingDirectory() error {
	r.in = &bytes.Buffer{}
	r.out = &bytes.Buffer{}
	r.err = nil
	r.urlList = urllist.New()
	dl, err := downloader.New(downloader.WithBaseDir(r.workingDir))
	if err != nil {
		return err
	}
	r.downloader = dl
	return nil
}

func (r *replFeature) theDownloadTargetDirectoryNameShouldBe(expected string) error {
	if r.downloader == nil {
		return fmt.Errorf("downloader not initialized")
	}
	if r.downloader.TargetName() != expected {
		return fmt.Errorf("expected target name %q, got %q", expected, r.downloader.TargetName())
	}
	return nil
}

func (r *replFeature) theDownloadTargetDirectoryShouldNotExist() error {
	if r.downloader == nil {
		return fmt.Errorf("downloader not initialized")
	}
	if _, err := os.Stat(r.downloader.TargetPath()); !os.IsNotExist(err) {
		return fmt.Errorf("expected target directory %s to not exist", r.downloader.TargetPath())
	}
	return nil
}

func (r *replFeature) theDownloadTargetDirectoryShouldExist() error {
	if r.downloader == nil {
		return fmt.Errorf("downloader not initialized")
	}
	info, err := os.Stat(r.downloader.TargetPath())
	if err != nil {
		return fmt.Errorf("target directory does not exist: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("target path is not a directory")
	}
	return nil
}

func (r *replFeature) ensureServer() {
	if r.server == nil {
		r.serverRoutes = make(map[string]string)
		r.serverStatuses = make(map[string]int)
		r.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if status, ok := r.serverStatuses[req.URL.Path]; ok && status != http.StatusOK {
				http.Error(w, http.StatusText(status), status)
				return
			}
			if body, ok := r.serverRoutes[req.URL.Path]; ok {
				fmt.Fprint(w, body)
				return
			}
			http.NotFound(w, req)
		}))
	}
}

func (r *replFeature) aTestWebServerServingAt(content, path string) error {
	r.ensureServer()
	r.serverRoutes[path] = content
	r.serverStatuses[path] = http.StatusOK
	return nil
}

func (r *replFeature) aTestWebServerRespondingWithAt(statusCode int, path string) error {
	r.ensureServer()
	r.serverStatuses[path] = statusCode
	return nil
}

func (r *replFeature) theURLListShouldNotContain(expected string) error {
	if r.server != nil {
		expected = strings.ReplaceAll(expected, "<server>", r.server.URL)
	}
	for _, u := range r.urlList.Get() {
		if u == expected {
			return fmt.Errorf("expected URL list to not contain %q, but found in %v", expected, r.urlList.Get())
		}
	}
	return nil
}

func (r *replFeature) theFileInTheTargetDirectoryShouldContain(fileName, expected string) error {
	filePath := filepath.Join(r.downloader.TargetPath(), fileName)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	if !strings.Contains(string(content), expected) {
		return fmt.Errorf("expected file %s to contain %q, got %q", fileName, expected, string(content))
	}
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	rf := &replFeature{}

	ctx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		rf.reset()
		return ctx, nil
	})

	ctx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		rf.reset()
		return ctx, nil
	})

	ctx.Step(`^the REPL application is initialized$`, rf.theREPLApplicationIsInitialized)
	ctx.Step(`^the REPL is started with no input$`, rf.theREPLIsStartedWithNoInput)
	ctx.Step(`^the user enters "([^"]*)"$`, rf.theUserEnters)
	ctx.Step(`^the user requests completion for "([^"]*)"$`, rf.theUserRequestsCompletionFor)
	ctx.Step(`^the output should contain "([^"]*)"$`, rf.theOutputShouldContain)
	ctx.Step(`^the output should not contain "([^"]*)"$`, rf.theOutputShouldNotContain)
	ctx.Step(`^the output should report unknown command "([^"]*)"$`, rf.theOutputShouldReportUnknownCommand)
	ctx.Step(`^the completion candidates should include "([^"]*)"$`, rf.theCompletionCandidatesShouldInclude)
	ctx.Step(`^the REPL session should terminate successfully$`, rf.theREPLSessionShouldTerminateSuccessfully)
	ctx.Step(`^a working directory with directory "([^"]*)"$`, rf.aWorkingDirectoryWithDirectory)
	ctx.Step(`^the REPL application is initialized in that working directory$`, rf.theREPLApplicationIsInitializedInThatWorkingDirectory)
	ctx.Step(`^the download target directory name should be "([^"]*)"$`, rf.theDownloadTargetDirectoryNameShouldBe)
	ctx.Step(`^the download target directory should not exist$`, rf.theDownloadTargetDirectoryShouldNotExist)
	ctx.Step(`^the download target directory should exist$`, rf.theDownloadTargetDirectoryShouldExist)
	ctx.Step(`^a test web server serving "([^"]*)" at "([^"]*)"$`, rf.aTestWebServerServingAt)
	ctx.Step(`^a test web server responding with (\d+) at "([^"]*)"$`, rf.aTestWebServerRespondingWithAt)
	ctx.Step(`^the URL list should not contain "([^"]*)"$`, rf.theURLListShouldNotContain)
	ctx.Step(`^the file "([^"]*)" in the target directory should contain "([^"]*)"$`, rf.theFileInTheTargetDirectoryShouldContain)
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"."},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
