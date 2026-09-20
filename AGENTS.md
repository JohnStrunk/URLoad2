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

## Testing standards

- High test coverage should be maintained (>90% statement coverage).
- Every EARS requirement must be verified by:
  - Gherkin scenarios executed with `godog`.
  - Property-based tests implemented with
    [`rapid`](https://pkg.go.dev/pgregory.net/rapid) to verify invariants
    across generated inputs with automated shrinking.
- Mutation testing should be performed using
  [`go-mutesting`](https://github.com/avito-tech/go-mutesting):
  - Run mutation tests using `go-mutesting ./internal/...` to verify test
    suite efficacy and kill logical mutants.

## Releases and distribution

- URLoad2 uses [GoReleaser](https://goreleaser.com) (`.goreleaser.yaml`) for
  multi-platform builds and packaging.
- Creating a release tag (e.g. `vX.Y.Z`) triggers the Release workflow
  (`.github/workflows/release.yaml`), attaching binary artifacts and archives
  to a draft GitHub release.

## Linting and formatting

- All files should be linted and formatted using `pre-commit`.
- Go code should be linted and formatted using `golangci-lint`.
