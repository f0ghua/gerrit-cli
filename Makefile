BINARY=gerrit
BUILD_DIR=bin

.PHONY: build install clean

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/gerrit

install:
	go install ./cmd/gerrit

clean:
	rm -rf $(BUILD_DIR)
