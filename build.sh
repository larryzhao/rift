BUILD=$(eval git rev-parse HEAD | awk '{ print substr($1,1,8) }')
VERSION=$(eval git describe --tags --match '*.*.*' --match '*.*.*-*' --match '*.*.*-rc.*' --first-parent --abbrev=0)
# go run -o ./rye -ldflags="-X main.Build=$BUILD -X main.Version=$VERSION" ./cmd/cli/...
go run -ldflags="-X main.Build=$BUILD -X main.Version=$VERSION" ./cmd/cli/... version