#!/bin/bash

go build . && mv -f xmux ~/.dotfiles/auto/os/Darwin/bin
GOOS=linux GOARCH=amd64 go build . && mv xmux ~/.dotfiles/auto/os/Linux/bin/
