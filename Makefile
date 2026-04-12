BINARY := personstorage
BIN_DIR := bin
BUILD_DIR := .cache/go-build

.PHONY: build test test-integration run clean

build:
	mkdir -p $(BIN_DIR) $(BUILD_DIR)
	GOCACHE=$(CURDIR)/$(BUILD_DIR) go build -o $(BIN_DIR)/$(BINARY) .

test:
	mkdir -p $(BUILD_DIR)
	GOCACHE=$(CURDIR)/$(BUILD_DIR) go test ./...

test-integration:
	mkdir -p $(BUILD_DIR)
	GOCACHE=$(CURDIR)/$(BUILD_DIR) go test -run TestPersonStorageIntegration .

run: build
	./$(BIN_DIR)/$(BINARY)

clean:
	rm -rf $(BIN_DIR) $(BUILD_DIR)
