# rift

Manage and run proxy connections from the command line.

## Install

Install with [Homebrew](https://brew.sh) (macOS):

```sh
brew tap larryzhao/rift
brew install rift
```

This installs a prebuilt binary for your architecture (Apple Silicon or Intel).
Upgrade later with:

```sh
brew upgrade rift
```

## Usage

```sh
rift --help
```

## Release

Releases are cut with [`just`](https://github.com/casey/just):

```sh
just release 1.2.3
```

This builds prebuilt `darwin/arm64` and `darwin/amd64` binaries, publishes them
to a GitHub release, and updates the Homebrew formula in the
`larryzhao/homebrew-rift` tap.
