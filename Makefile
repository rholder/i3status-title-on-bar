BUILD_DIR=build

NAME=i3status-title-on-bar
VERSION=0.5.0-dev

.PHONY: all build coverage clean fmt test tree

all: clean build

clean:
	rm -rfv $(BUILD_DIR)

build:
	mkdir -p $(BUILD_DIR)
	cd cmd/$(NAME)/; go build -ldflags "-w -s -X main.Version=$(VERSION)" -o ../../$(BUILD_DIR)/$(NAME)
	@echo Build successful.

fmt:
	gofmt -w ./cmd/ ./pkg/

test:
	mkdir -p $(BUILD_DIR)
	go test -coverprofile $(BUILD_DIR)/coverage.out ./cmd/... ./pkg/...

coverage: test
	go tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html

tree:
	tree -I tmp --matchdirs
