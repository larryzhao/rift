BUILD=$(eval git rev-parse HEAD)
VERSION=$(eval git describe --tags $(git rev-list --tags --max-count=1))
mkdir -p ./dist >/dev/null 2>&1
go build -o ./dist/rift-$VERSION-macos-arm64/rift -tags "with_quic,with_utls,with_grpc,with_wireguard,with_dhcp,with_acme" -ldflags "-s -w -X github.com/larryzhao/rift.Build=$BUILD -X github.com/larryzhao/rift.Version=$VERSION" ./cmd/cli/...
touch ./dist/rift-$VERSION-macos-arm64/LICENSE
touch ./dist/rift-$VERSION-macos-arm64/README.md

# echo "BUILD: $BUILD"
# echo "VERSION: $VERSION"
# go run -ldflags "-s -w -X github.com/larryzhao/rift.Build=$BUILD -X github.com/larryzhao/rift.Version=$VERSION" ./cmd/cli/... version