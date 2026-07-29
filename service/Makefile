.PHONY: build test lint run

build:
	go build -ldflags "-X faridoon/service/buildinfo.Version=$$(cat ../VERSION 2>/dev/null || echo development)" -o faridoon-service .

test:
	go test ./...

lint:
	go vet ./...

run:
	go run .
