.PHONY: build buf

build:
	go build -o bin/butter ./cmd/butter

buf:
	buf generate
	protoc-go-inject-tag -input="pkg/proto/agents/v1/*.pb.go"
