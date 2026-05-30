branch := "main"
os := "darwin"
arch := "amd64"

# List available recipes
default:
    @just --list

# Release a new version, e.g. `just release 1.2.3`
release version:
    #!/usr/bin/env bash
    set -euo pipefail

    version="{{ version }}"
    if [[ -z "$version" ]]; then
        echo "err: version not specified"
        exit 1
    fi
    # validate semver-ish version (digits separated by dots, optional pre-release/build)
    if ! [[ "$version" =~ ^[0-9]+(\.[0-9]+)*([.-].+)?$ ]]; then
        echo "err: invalid version: $version"
        exit 1
    fi

    # check we're on the release branch
    if [[ "$(git branch --show-current)" != "{{ branch }}" ]]; then
        echo "err: branch is not {{ branch }}"
        exit 1
    fi

    # check there are no uncommitted changes
    if [[ -n "$(git status --porcelain)" ]]; then
        echo "err: unstaged changes found"
        exit 1
    fi

    dist_name="rye-${version}-{{ os }}-{{ arch }}"
    dist_dir="./dist/${dist_name}"

    # build binary
    mkdir -p "$dist_dir"
    build="$(git rev-parse HEAD)"
    GOOS="{{ os }}" GOARCH="{{ arch }}" go build \
        -o "$dist_dir/rye" \
        -ldflags "-s -w -X github.com/larryzhao/rye.Build=${build} -X github.com/larryzhao/rye.Version=${version}" \
        ./cmd/cli/...
    touch "$dist_dir/LICENSE"
    touch "$dist_dir/README.md"

    # package
    tar czf "dist/${dist_name}.tar.gz" -C dist "$dist_name"

    # create the release and upload the artifact
    gh release create "v${version}"
    gh release upload "v${version}" "./dist/${dist_name}.tar.gz"
