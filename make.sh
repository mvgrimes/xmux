#!/bin/bash

# go build . && mv -f xmux ~/.dotfiles/auto/host/Darwin/bin
GOOS=linux GOARCH=amd64 go build . && mv xmux ~/.dotfiles/auto/os/Linux/bin/
GOOS=darwin GOARCH=arm64 go build . && mv xmux ~/.dotfiles/auto/os-arch/Darwin-arm64/bin/
GOOS=darwin GOARCH=amd64 go build . && mv xmux ~/.dotfiles/auto/os-arch/Darwin-x86_64/bin/
