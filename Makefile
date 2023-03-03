BIN_BRUTE_FP=./bin/brutefp
BRUTE_FP_CFG_FILE=./deployments/configs/brutefp_config.json
BRUTE_FP_CFG_TPL=./deployments/configs/brutefp_config.json.template
BIN_BRUTE_FP_CLI=./bin/brutefp-cli
ENV_NAME=tests

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

up: cfg-up
	env `cat ./deployments/env.${ENV_NAME}` docker-compose -f ./deployments/docker-compose.yaml up -d --build

down: cfg-clean
	env `cat ./deployments/env.${ENV_NAME}` docker-compose -f ./deployments/docker-compose.yaml down

restart: down up

cfg-up:
	env `cat ./deployments/env.${ENV_NAME}` envsubst < ${BRUTE_FP_CFG_TPL} > ${BRUTE_FP_CFG_FILE};

cfg-clean:
	rm -f ${BRUTE_FP_CFG_FILE}

integration-test: cfg-up
	set -e ;\
	env `cat ./deployments/env.${ENV_NAME}` docker-compose -f ./deployments/docker-compose.yaml up -d --build;\
	test_status_code=0 ;\
	docker build -t tests --network host -f tests/Dockerfile . || test_status_code=$$? ;\
	env `cat ./deployments/env.${ENV_NAME}` docker-compose -f ./deployments/docker-compose.yaml down --rmi local --volumes --remove-orphans; \
    rm -f ${BRUTE_FP_CFG_FILE}; \
	exit $$test_status_code ;

integration-test-cleanup: cfg-clean
	env `cat ./deployments/env.${ENV_NAME}` docker-compose -f ./deployments/docker-compose.yaml down --rmi local --volumes --remove-orphans;

build:
	go build -v -o $(BIN_BRUTE_FP) -ldflags "$(LDFLAGS)" ./cmd/brutefp
	go build -v -o $(BIN_BRUTE_FP_CLI) -ldflags "$(LDFLAGS)" ./cmd/brutefp-cli

run: build
	$(BIN_BRUTE_FP) -config ./configs/brutefp_config.json

version: build
	$(BIN_BRUTE_FP) version

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