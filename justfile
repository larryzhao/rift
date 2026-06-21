branch := "main"

# macOS architectures to build prebuilt binaries for
archs := "arm64 amd64"

# sing-box feature build tags (keep in sync with build.sh)
build_tags := "with_quic,with_utls,with_grpc,with_wireguard,with_dhcp,with_acme"

# Homebrew tap repo that hosts the generated formula
tap_repo := "git@github.com:larryzhao/homebrew-rift.git"

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

    build="$(git rev-parse HEAD)"
    rm -rf dist
    mkdir -p dist

    # build + package one tarball per macOS architecture
    # (bash 3.2 on macOS has no associative arrays, so use eval'd sha_<arch> vars)
    assets=()
    for goarch in {{ archs }}; do
        dist_name="rift-${version}-darwin-${goarch}"
        dist_dir="dist/${dist_name}"
        mkdir -p "$dist_dir"

        GOOS=darwin GOARCH="$goarch" go build \
            -tags "{{ build_tags }}" \
            -o "$dist_dir/rift" \
            -ldflags "-s -w -X github.com/larryzhao/rift.Build=${build} -X github.com/larryzhao/rift.Version=${version}" \
            ./cmd/cli/...
        cp LICENSE README.md "$dist_dir/"

        # archive files at the root of the tarball (no wrapping dir)
        tar czf "dist/${dist_name}.tar.gz" -C "$dist_dir" .
        eval "sha_${goarch}=\"\$(shasum -a 256 'dist/${dist_name}.tar.gz' | awk '{print \$1}')\""
        assets+=("dist/${dist_name}.tar.gz")
    done

    # create the GitHub release and upload every artifact
    gh release create "v${version}" --title "v${version}" --generate-notes "${assets[@]}"

    # render and publish the Homebrew formula into the tap repo
    base_url="https://github.com/larryzhao/rift/releases/download/v${version}"
    tap_dir="dist/homebrew-rift"
    git clone "{{ tap_repo }}" "$tap_dir"
    mkdir -p "$tap_dir/Formula"
    cat > "$tap_dir/Formula/rift.rb" <<EOF
    class Rift < Formula
      desc "Manage and run proxy connections from the command line"
      homepage "https://github.com/larryzhao/rift"
      version "${version}"
      license "GPL-3.0-or-later"

      on_macos do
        on_arm do
          url "${base_url}/rift-${version}-darwin-arm64.tar.gz"
          sha256 "${sha_arm64}"
        end
        on_intel do
          url "${base_url}/rift-${version}-darwin-amd64.tar.gz"
          sha256 "${sha_amd64}"
        end
      end

      def install
        bin.install "rift"
      end

      test do
        system "#{bin}/rift", "--help"
      end
    end
    EOF

    git -C "$tap_dir" checkout -B main
    git -C "$tap_dir" add Formula/rift.rb
    git -C "$tap_dir" commit -m "rift ${version}"
    git -C "$tap_dir" push -u origin main

    echo "released v${version} and updated tap formula"
