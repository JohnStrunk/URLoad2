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

- `help` or `?`: Display available commands
- `version`: Display version information
- `exit` or `quit`: Terminate the REPL session

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
