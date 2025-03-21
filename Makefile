.PHONY: test lint quality fmt

BINARY_NAME=yadfls
VERSION=0.0.1

COMMIT_HASH=$(shell git rev-parse --short HEAD)
BUILD_TIMESTAMP=$(shell date '+%Y-%m-%dT%H:%M:%S')

LDFLAGS=-ldflags "-X 'main.Version=${VERSION}' -X 'main.CommitHash=${COMMIT_HASH}' -X 'main.BuildTimestamp=${BUILD_TIMESTAMP}'"

LINUX_AMD64 = build/${BINARY_NAME}_linux_amd64

build: ${LINUX_AMD64}

${LINUX_AMD64}:
	GOARCH=amd64 GOOS=linux go build ${LDFLAGS} -o ${LINUX_AMD64} cmd/yadlfs/yadlfs.go

test: ## run tests
	go test -timeout 30s -race -failfast ./...

lint: ## run server linting
	golangci-lint run --fix

fmt: ## run server prettifier
	go fmt ./...

quality: fmt lint ## check-quality