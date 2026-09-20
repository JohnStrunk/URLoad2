# URLoad2

URLoad2 is a statically linked Go REPL (Read-Eval-Print Loop) application.

## Overview

- **Language:** Go
- **Binary:** Statically linked executable (`CGO_ENABLED=0`)
- **Interactive:** Terminal line editing with tab completion support
- **Spec-Driven:** Functionality specified via EARS (Easy Approach to
  Requirements Syntax) in Gherkin feature files executed with `godog`
- **Quality & CI:** Automated linting via `pre-commit` and `golangci-lint`, with
  automated testing and building in GitHub Actions CI

## Building

To build the statically linked binary:

```bash
CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/urload2 ./cmd/urload2
```

## Running

Run the interactive REPL:

```bash
./bin/urload2
```

Available commands within the REPL:

- `add <url>`: Append a URL to the end of the list
- `get`: Download each URL in the list, saving it as a file in the session
  target directory
- `head <n>`: Keep the first n URLs in the list, discarding the rest
- `tail <n>`: Keep the last n URLs in the list, discarding the rest
- `list`: Display the current list of URLs
- `clear`: Clear the URL list
- `help` or `?`: Display available commands
- `version`: Display version information
- `exit` or `quit`: Terminate the REPL session

The REPL prompt dynamically displays the target directory and current
length of the URL list (e.g., `URLoad2 [0000] (0)>`).
On startup, the REPL determines its download target directory (a 4-digit
numbered subdirectory of cwd, e.g. `0000`). The target directory is created
upon the first `get` command.
The REPL supports tab completion for all commands.

## Testing

Run all unit tests and Godog feature specifications:

```bash
go test -v ./...
```

## Linting and Code Quality

Format and lint files using `pre-commit` and `golangci-lint`:

```bash
pre-commit run --all-files
golangci-lint run ./...
```
