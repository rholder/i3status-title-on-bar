.PHONY: all build coverage clean fmt test

NAME = i3status-title-on-bar
BIN_NAME = $(NAME)
VERSION = $(shell git describe --tags --always --dirty)

BUILD_DIR = build

all: clean build

## clean: remove the build directory
clean:
	rm -rfv $(BUILD_DIR)

## build: assemble the project and place a binary in build/ for this OS
build:
	mkdir -p $(BUILD_DIR)
	go build -ldflags "-w -s -X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BIN_NAME) ./cmd/$(NAME)/main.go
	@echo Build successful.

## fmt: run gofmt for the project
fmt:
	gofmt -w ./cmd/ ./pkg/

## test: run the unit tests
test:
	mkdir -p $(BUILD_DIR)
	go test -v -coverprofile $(BUILD_DIR)/coverage.out ./cmd/... ./pkg/...

## coverage: run the unit tests with test coverage output to build/coverage.html
coverage: test
	go tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html

## help: prints this help message
help:
	@echo "Usage:\n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
