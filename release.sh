BUILD=$(eval git rev-parse HEAD)
VERSION=$(eval git describe --tags $(git rev-list --tags --max-count=1))
mkdir -p ./dist >/dev/null 2>&1
go build -o ./dist/rye-$VERSION-macos-arm64/rye -ldflags "-s -w -X github.com/larryzhao/rye.Build=$BUILD -X github.com/larryzhao/rye.Version=$VERSION" ./cmd/cli/...
touch ./dist/rye-$VERSION-macos-arm64/LICENSE
touch ./dist/rye-$VERSION-macos-arm64/README.md

# echo "BUILD: $BUILD"
# echo "VERSION: $VERSION"
# go run -ldflags "-s -w -X github.com/larryzhao/rye.Build=$BUILD -X github.com/larryzhao/rye.Version=$VERSION" ./cmd/cli/... version