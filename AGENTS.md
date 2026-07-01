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
- `just release <version>`: cut a tagged release (e.g. `just release 1.2.3`). See **Releasing** below.

For quick iteration, run `go test ./... && go build ./cmd/cli/...` before opening a PR.

## Releasing
`just release <version>` (e.g. `just release 0.8.0`) is the single entry point. It requires a **clean working tree on `main`**, and `main` must already be pushed to the remote (the `v<version>` tag is created at the current `HEAD`, so land your commit on remote `main` first). The recipe:
1. Builds `darwin/arm64` and `darwin/amd64` with the sing-box `build_tags` and `Build`/`Version` ldflags.
2. Packages `dist/rift-<version>-darwin-<arch>.tar.gz` (LICENSE + README + `rift` at the tarball root).
3. `gh release create v<version> --generate-notes` with both tarballs (over HTTPS).
4. Clones the Homebrew tap (`homebrew-rift`), rewrites `Formula/rift.rb` with the new version + per-arch sha256, commits, and pushes.

**No-SSH environments:** the tap step defaults to the SSH remote (`git@github.com:larryzhao/homebrew-rift.git`). If SSH to GitHub is blocked but `gh` (HTTPS) works, override the tap remote so the whole flow runs over HTTPS — `gh` provides git credentials via its credential helper:
```
just tap_repo='https://github.com/larryzhao/homebrew-rift.git' release <version>
```
Push `main` over HTTPS the same way if needed: `git push https://github.com/larryzhao/rift.git main`.

**macOS local reinstall gotcha:** overwriting an existing binary in place with `cp` (e.g. into `~/.local/bin/rift`) can make the next `exec` die with `SIGKILL` (exit 137) due to a stale code-signing cache/cdhash. Fix by removing the old file before copying (`rm -f <dest> && cp ...`), or re-sign after copying: `codesign -f -s - <dest>`.

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
