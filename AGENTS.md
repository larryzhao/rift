# Repository Guidelines

## Project Structure & Module Organization
This repository is a Go CLI project. Keep new code organized by runtime concern:
- `cmd/cli/main.go`: primary CLI entrypoint.
- `commands/`: Cobra subcommands (`start`, `stop`, `status`, `servers`, etc.).
- `xray/`, `hysteria2/`, `pac/`, `runner/`: backend-specific runners/configuration.
- Root `*.go` files: shared domain logic (server, repository, subscription, status).
- `.github/workflows/go.yml`: CI/release automation.
- `build.sh`: release packaging script that writes binaries to `dist/`.

If you add a new feature flag or provider, expose it through `commands/` and keep provider-specific logic in its own package.

## Build, Test, and Development Commands
Use these commands from the repository root:
- `go run ./cmd/cli/...`: run the CLI locally.
- `go build ./cmd/cli/...`: compile CLI packages and catch build errors.
- `go test ./...`: run all unit tests (add tests with new behavior changes).
- `./build.sh`: produce a release binary in `dist/` with embedded `Build`/`Version` ldflags.

For quick iteration, run `go test ./... && go build ./cmd/cli/...` before opening a PR.

## Coding Style & Naming Conventions
- Follow idiomatic Go style and always run `gofmt` on edited files.
- Use short, clear package names (lowercase, no underscores).
- Exported identifiers use `CamelCase`; internal helpers use `camelCase`.
- Prefer small, focused functions in `commands/` and keep command wiring separate from business logic.
- Keep file names descriptive, e.g., `commands/connect.go`, `hysteria2/config.go`.

## Testing Guidelines
- Place tests next to implementation files using `*_test.go` naming.
- Prefer table-driven tests for command/config parsing logic.
- Cover new branches and failure paths (invalid config, missing server, runner errors).
- Run `go test ./...` locally before pushing.

## Commit & Pull Request Guidelines
Git history mixes conventional and ad-hoc messages, but `feat:`/`fix:` prefixes are present and preferred.
- Use concise, imperative commit titles, ideally with prefixes (`feat:`, `fix:`, `refactor:`, `chore:`).
- Keep commits scoped to a single concern.
- PRs should include: summary, motivation, test evidence (`go test`/`go build` output), and linked issue (if applicable).
- For CLI behavior changes, include sample command/output snippets.
