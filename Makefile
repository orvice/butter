.PHONY: build butter butter-daemon buf

build: butter butter-daemon

butter:
	go build -o bin/butter ./cmd/butter

butter-daemon:
	go build -o bin/butter-daemon ./cmd/butter-daemon

buf:
	buf generate
	protoc-go-inject-tag -input="pkg/proto/agents/v1/*.pb.go"
