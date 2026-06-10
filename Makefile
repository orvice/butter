.PHONY: build butter butter-daemon buf

DAEMON_COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo unknown)
DAEMON_LDFLAGS := -X main.daemonCommit=$(DAEMON_COMMIT)

build: butter butter-daemon

butter:
	go build -o bin/butter ./cmd/butter

butter-daemon:
	go build -ldflags "$(DAEMON_LDFLAGS)" -o bin/butter-daemon ./cmd/butter-daemon

buf:
	buf generate
	protoc-go-inject-tag -input="pkg/proto/agents/v1/*.pb.go"
