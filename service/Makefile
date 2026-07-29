.PHONY: build test lint run codestyle go-tools

build:
	go build -ldflags "-X faridoon/service/buildinfo.Version=$$(cat ../VERSION 2>/dev/null || echo development)" -o faridoon-service .

test:
	go test ./...

lint: codestyle

codestyle: go-tools
	go fmt ./...
	go vet ./...
	golangci-lint run ./...

go-tools:
	go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2"

run:
	go run .
