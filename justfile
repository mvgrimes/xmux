APP       := "xmux"
VER_FILE  := "./cmd/xmux/main.go"
MAIN_FILE := "./cmd/xmux/"
VERSION   := shell('perl -nE "m{version\\s*=\\s*\"(\\d+\\.\\d+\\.\\d+)\"}i && print \$1" ' + VER_FILE)

build:
  echo "Building version {{VERSION}} of {{APP}}"
  go build -o {{APP}} {{MAIN_FILE}}

lint:
  go vet ./... || true
  golangci-lint run ./... || true
  govulncheck ./...

fmt:
  go fmt ./...

test:
  # go test ./...
  gotestsum

release:
  go mod tidy
  just fmt
  just build
  git diff --exit-code
  git tag "{{VERSION}}"
  git push
  git release
  git push --tags
  goreleaser release --clean
