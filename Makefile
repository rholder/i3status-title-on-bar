BUILD_DIR=build

NAME=i3status-custom-bar
VERSION=0.2.0

.PHONY: all build clean fmt test

all: clean build

clean:
	rm -rfv $(BUILD_DIR)

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(NAME) $(NAME)
	@echo Build successful.

fmt:
	go fmt ./src/...

test:
	@echo "TODO create test Makefile target"
