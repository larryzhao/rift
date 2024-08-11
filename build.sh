BUILD=$(eval git rev-parse HEAD)
VERSION=$(eval git describe --tags $(git rev-list --tags --max-count=1))
go build -o ./rye -ldflags "-s -w -X github.com/larryzhao/rye.Build=$BUILD -X github.com/larryzhao/rye.Version=$VERSION" ./cmd/cli/...

# echo "BUILD: $BUILD"
# echo "VERSION: $VERSION"
# go run -ldflags "-s -w -X github.com/larryzhao/rye.Build=$BUILD -X github.com/larryzhao/rye.Version=$VERSION" ./cmd/cli/... version