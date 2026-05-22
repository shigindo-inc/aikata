# aikata Makefile
#
# Common targets:
#   make build    — build the aikata binary into ./aikata
#   make test     — run all Go tests
#   make lint     — run golangci-lint
#   make install  — install aikata into $GOBIN, or $GOPATH/bin when unset
#   make run      — go run aikata (pass flags via ARGS="--help")
#   make clean    — remove build artifacts
#   make tidy     — go mod tidy
#   make verify   — full pre-commit gate: tidy + test + lint + build

BINARY  := aikata
PKG     := github.com/shigindo-inc/aikata
VERSION ?= $(shell git describe --tags --dirty 2>/dev/null || echo "0.0.1-dev")
LDFLAGS := -X main.version=$(VERSION)
ARGS    ?=

.PHONY: build
build:
	go build -ldflags '$(LDFLAGS)' -o ./$(BINARY) ./cmd/aikata

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: install
install:
	go install -ldflags '$(LDFLAGS)' ./cmd/aikata

.PHONY: run
run:
	go run ./cmd/aikata $(ARGS)

.PHONY: clean
clean:
	rm -f $(BINARY)

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: update-golden
update-golden:
	go test ./internal/scaffold/... -run TestGolden_ -update

.PHONY: verify
verify: tidy test lint build
