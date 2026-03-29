APP       := "xmux"
VER_FILE  := "./main.go"
MAIN_FILE := "./main.go"
VERSION   := shell('perl -nE "m{version\\s*=\\s*\"(\\d+\\.\\d+\\.\\d+)\"}i && print \$1" ' + VER_FILE)

build:
  echo "Building {{APP}}"
  go build -o {{APP}} {{MAIN_FILE}}

lint:
  go vet ./... || true
  golangci-lint run ./... || true
  govulncheck ./...

fmt:
  go fmt ./...

test:
  go test ./...

install:
  GOOS=linux  GOARCH=amd64  go build -o {{APP}} {{MAIN_FILE}} && mv {{APP}} ~/.dotfiles/auto/os/Linux/bin/
  GOOS=darwin GOARCH=arm64  go build -o {{APP}} {{MAIN_FILE}} && mv {{APP}} ~/.dotfiles/auto/os-arch/Darwin-arm64/bin/
  GOOS=darwin GOARCH=amd64  go build -o {{APP}} {{MAIN_FILE}} && mv {{APP}} ~/.dotfiles/auto/os-arch/Darwin-x86_64/bin/

release:
  go mod tidy
  just fmt
  just build
  git diff --exit-code
  git tag "{{VERSION}}"
  git push
  git push --tags
  goreleaser release --clean
