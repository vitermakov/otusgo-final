BIN_APP=./bin/brutefp

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	go build -v -o $(BIN_APP) -ldflags "$(LDFLAGS)" ./cmd/brutefp

run: build
	$(BIN_APP) -config ./configs/brutefp_config.json

version: build
	$(BIN_APP) version

test:
	go test -race ./internal/... ./pkg/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.50.1

lint: install-lint-deps
	golangci-lint run ./...

generate:
	rm -Rf internal/handler/grpc/pb
	mkdir internal/handler/grpc/pb
	protoc --proto_path=internal/handler/grpc/proto/ --go_out=. --go-grpc_out=. internal/handler/grpc/proto/*.proto