.PHONY: test vet vet-darwin build run tidy

test:
	go test ./...

vet:
	go vet ./...

# The window code is macOS-only; this type-checks it from any host.
vet-darwin:
	GOOS=darwin CGO_ENABLED=0 go vet ./...

build:
	go build -o bin/butter-desktop ./cmd/butter-desktop

run:
	go run ./cmd/butter-desktop

tidy:
	go mod tidy
