.DEFAULT_GOAL := build

.PHONY: fmt lint vet build

fmt:
	go fmt ./...

lint: fmt
	staticcheck ./...

vet: fmt
	go vet ./...

build: vet
	go build -o ./bin/mydb ./cmd

run: vet
	go run ./cmd/main.go

clean:
	rm -rf ./bin
