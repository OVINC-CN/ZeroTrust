.PHONY: build run clean test deps

BINARY_NAME=zerotrust
BUILD_DIR=bin

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/zerotrust

run: build
	./$(BUILD_DIR)/$(BINARY_NAME) -config configs/config.yaml

clean:
	rm -rf $(BUILD_DIR)

test:
	go test -v ./...

deps:
	go mod tidy
	go mod download

docker-build:
	docker build -t $(BINARY_NAME):latest .
