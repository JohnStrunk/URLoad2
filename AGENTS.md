# AGENTS.md

## General

- Development should be done on a separate branch, using worktrees.
- Worktrees can be created in the `.worktrees` directory.

## Spec-driven development

- All development, configuration, and system behavior is driven and defined by
  specifications.
- Specficiations are written using EARS (Easy Approach to Requirements Syntax)
  format.
- Gherkin is the Domain Specific Language (DSL) used to write the specs.
  - Functionality is defined in Gherkin `.feature` files.
  - The EARS Specifications are written in the `Rules` lines of Gherkin feature
    files.

## Development environment

- The primary language for URLoad2 is Golang (go).
- Gherkin files should be executed using
  [`godog`](https://github.com/cucumber/godog).

## Linting and formatting

- All files should be linted and formatted using `pre-commit`.
- Go code should be linted and formatted using `golangci-lint`.
