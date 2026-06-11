.PHONY: build butter butter-daemon buf

BUILD_COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo unknown)
SERVER_LDFLAGS := -X main.serverCommit=$(BUILD_COMMIT)
DAEMON_LDFLAGS := -X main.daemonCommit=$(BUILD_COMMIT)

build: butter butter-daemon

butter:
	go build -ldflags "$(SERVER_LDFLAGS)" -o bin/butter ./cmd/butter

butter-daemon:
	go build -ldflags "$(DAEMON_LDFLAGS)" -o bin/butter-daemon ./cmd/butter-daemon

buf:
	buf generate
	protoc-go-inject-tag -input="pkg/proto/agents/v1/*.pb.go"
