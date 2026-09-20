package features_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"github.com/JohnStrunk/URLoad2/internal/repl"
)

type replFeature struct {
	in  *bytes.Buffer
	out *bytes.Buffer
	err error
}

func (r *replFeature) theREPLApplicationIsInitialized() error {
	r.in = &bytes.Buffer{}
	r.out = &bytes.Buffer{}
	r.err = nil
	return nil
}

func (r *replFeature) theREPLIsStartedWithNoInput() error {
	app := repl.New(r.in, r.out)
	r.err = app.Run(context.Background())
	return nil
}

func (r *replFeature) theUserEnters(input string) error {
	r.in.WriteString(input + "\n")
	app := repl.New(r.in, r.out)
	r.err = app.Run(context.Background())
	return nil
}

func (r *replFeature) theOutputShouldContain(expected string) error {
	actual := r.out.String()
	if !strings.Contains(actual, expected) {
		return fmt.Errorf("expected output to contain %q, but got %q", expected, actual)
	}
	return nil
}

func (r *replFeature) theOutputShouldReportUnknownCommand(cmd string) error {
	expected := fmt.Sprintf("unknown command: %q", cmd)
	return r.theOutputShouldContain(expected)
}

func (r *replFeature) theREPLSessionShouldTerminateSuccessfully() error {
	if r.err != nil {
		return fmt.Errorf("expected clean termination, but got error: %w", r.err)
	}
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	rf := &replFeature{}

	ctx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		return ctx, rf.theREPLApplicationIsInitialized()
	})

	ctx.Step(`^the REPL application is initialized$`, rf.theREPLApplicationIsInitialized)
	ctx.Step(`^the REPL is started with no input$`, rf.theREPLIsStartedWithNoInput)
	ctx.Step(`^the user enters "([^"]*)"$`, rf.theUserEnters)
	ctx.Step(`^the output should contain "([^"]*)"$`, rf.theOutputShouldContain)
	ctx.Step(`^the output should report unknown command "([^"]*)"$`, rf.theOutputShouldReportUnknownCommand)
	ctx.Step(`^the REPL session should terminate successfully$`, rf.theREPLSessionShouldTerminateSuccessfully)
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
