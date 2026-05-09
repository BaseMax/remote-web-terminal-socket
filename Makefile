BINARY   := remote-web-terminal-socket
BUILD_DIR := bin

.PHONY: all build run hash clean

all: build

build:
	go build -o $(BUILD_DIR)/$(BINARY) .

run: build
	$(BUILD_DIR)/$(BINARY)

## Usage: make hash PASS=mysecret
hash:
	@go run tools/hashpw/main.go "$(PASS)"

tidy:
	go mod tidy

clean:
	rm -rf $(BUILD_DIR)
