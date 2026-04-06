#!/bin/bash

VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS="-X main.version=${VERSION}"

# go build . && mv -f xmux ~/.dotfiles/auto/host/Darwin/bin
GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" ./cmd/xmux && mv xmux ~/.dotfiles/auto/os/Linux/bin/
GOOS=darwin GOARCH=arm64 go build -ldflags "${LDFLAGS}" ./cmd/xmux && mv xmux ~/.dotfiles/auto/os-arch/Darwin-arm64/bin/
GOOS=darwin GOARCH=amd64 go build -ldflags "${LDFLAGS}" ./cmd/xmux && mv xmux ~/.dotfiles/auto/os-arch/Darwin-x86_64/bin/
