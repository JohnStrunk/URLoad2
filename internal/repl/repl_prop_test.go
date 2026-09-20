package repl_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"pgregory.net/rapid"

	"github.com/JohnStrunk/URLoad2/internal/repl"
)

// EARS Rule 1: When the REPL starts, the REPL shall display an interactive prompt.
func TestProperty_EARS_PromptOnStart(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		prompt := rapid.StringMatching(`[a-zA-Z0-9_\-><#$ ]{1,30}`).Draw(t, "prompt")

		var out bytes.Buffer
		r := repl.New(strings.NewReader(""), &out, repl.WithPrompt(prompt))
		err := r.Run(context.Background())
		if err != nil {
			t.Fatalf("expected nil error on start with EOF, got %v", err)
		}

		actual := out.String()
		if !strings.HasPrefix(actual, prompt) {
			t.Fatalf("expected output to start with prompt %q, got %q", prompt, actual)
		}
	})
}

// EARS Rule 2: While running, when the user inputs the exit command, the REPL shall terminate gracefully.
func TestProperty_EARS_ExitCommandTerminatesGracefully(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		exitCmd := rapid.SampledFrom([]string{"exit", "quit"}).Draw(t, "exitCmd")
		leadingSpaces := rapid.StringMatching(`[ \t]*`).Draw(t, "leadingSpaces")
		trailingSpaces := rapid.StringMatching(`[ \t]*`).Draw(t, "trailingSpaces")

		input := leadingSpaces + exitCmd + trailingSpaces + "\n"
		var out bytes.Buffer
		r := repl.New(strings.NewReader(input), &out)
		err := r.Run(context.Background())
		if err != nil {
			t.Fatalf("expected nil error on graceful exit, got %v", err)
		}

		actual := out.String()
		if !strings.Contains(actual, "Goodbye!") {
			t.Fatalf("expected output to contain 'Goodbye!', got %q", actual)
		}
	})
}

// EARS Rule 3 & 4: While running, when the user inputs the help [or ?] command, the REPL shall display available commands.
func TestProperty_EARS_HelpDisplaysAvailableCommands(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		cmd := rapid.SampledFrom([]string{"help", "?"}).Draw(t, "cmd")
		leadingSpaces := rapid.StringMatching(`[ \t]*`).Draw(t, "leadingSpaces")
		trailingSpaces := rapid.StringMatching(`[ \t]*`).Draw(t, "trailingSpaces")

		input := leadingSpaces + cmd + trailingSpaces + "\nexit\n"
		var out bytes.Buffer
		r := repl.New(strings.NewReader(input), &out)
		err := r.Run(context.Background())
		if err != nil {
			t.Fatalf("expected nil error running help command, got %v", err)
		}

		actual := out.String()
		if !strings.Contains(actual, "Available commands:") {
			t.Fatalf("expected output to contain 'Available commands:', got %q", actual)
		}

		for _, availableCmd := range repl.Commands() {
			if !strings.Contains(actual, availableCmd) {
				t.Fatalf("expected output to contain command %q, got %q", availableCmd, actual)
			}
		}
	})
}

// EARS Rule 5: While the user is entering input, when completion is requested, the REPL shall return matching command candidates.
func TestProperty_EARS_CompletionMatchesCandidates(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		prefix := rapid.StringMatching(`[a-zA-Z0-9?_-]{0,10}`).Draw(t, "prefix")
		candidates := repl.Complete(prefix)

		seen := make(map[string]bool)
		for _, c := range candidates {
			if !strings.HasPrefix(c, prefix) {
				t.Fatalf("candidate %q does not start with prefix %q", c, prefix)
			}
			if seen[c] {
				t.Fatalf("duplicate candidate found: %q", c)
			}
			seen[c] = true
		}

		for _, availableCmd := range repl.Commands() {
			if strings.HasPrefix(availableCmd, prefix) {
				if !seen[availableCmd] {
					t.Fatalf("expected matching command %q for prefix %q to be in candidates %v", availableCmd, prefix, candidates)
				}
			}
		}

		if prefix == "" {
			if len(candidates) != len(repl.Commands()) {
				t.Fatalf("empty prefix expected all commands, got %v", candidates)
			}
		}
	})
}

// EARS Rule 6: If an unrecognized command is entered, then the REPL shall display an error message and continue running.
func TestProperty_EARS_UnrecognizedCommandReportsErrorAndContinues(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		validCmds := make(map[string]bool)
		for _, c := range repl.Commands() {
			validCmds[c] = true
		}

		unknownCmd := rapid.StringMatching(`[a-zA-Z0-9_]{1,15}`).
			Filter(func(s string) bool { return !validCmds[s] }).
			Draw(t, "unknownCmd")

		input := unknownCmd + "\nexit\n"
		var out bytes.Buffer
		r := repl.New(strings.NewReader(input), &out)
		err := r.Run(context.Background())
		if err != nil {
			t.Fatalf("expected nil error when recovering from unrecognized command, got %v", err)
		}

		actual := out.String()
		expectedErrorMsg := fmt.Sprintf("unknown command: %q", unknownCmd)
		if !strings.Contains(actual, expectedErrorMsg) {
			t.Fatalf("expected output to contain error %q, got %q", expectedErrorMsg, actual)
		}

		if !strings.Contains(actual, "Goodbye!") {
			t.Fatalf("expected REPL to continue running until exit, but 'Goodbye!' missing from output: %q", actual)
		}
	})
}

// EARS: While running, when displaying the command prompt, the REPL shall include the current length of the list.
func TestProperty_EARS_PromptReflectsListLength(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		numAdds := rapid.IntRange(0, 10).Draw(rt, "numAdds")
		var input strings.Builder
		for i := 0; i < numAdds; i++ {
			fmt.Fprintf(&input, "add http://example.com/%d\n", i)
		}
		input.WriteString("exit\n")

		var out bytes.Buffer
		r := repl.New(strings.NewReader(input.String()), &out)
		err := r.Run(context.Background())
		if err != nil {
			t.Fatalf("unexpected error running REPL: %v", err)
		}

		actual := out.String()
		for i := 0; i <= numAdds; i++ {
			expectedPrompt := fmt.Sprintf("urload2 [%d]> ", i)
			if !strings.Contains(actual, expectedPrompt) {
				t.Fatalf("expected output to contain prompt %q, got %q", expectedPrompt, actual)
			}
		}
	})
}
