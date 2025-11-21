.PHONY: build build-all clean test run dev install-node build-node

# Go binary name
BINARY=export-engine
GO_MODULE=github.com/turbo-export-engine

# Build for current platform
build:
	go build -o bin/$(BINARY) ./cmd/export-engine

# Build for all platforms
build-all: build-macos build-linux build-windows

build-macos:
	GOOS=darwin GOARCH=amd64 go build -o bin/$(BINARY)-macos ./cmd/export-engine

build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY)-linux ./cmd/export-engine

build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/$(BINARY)-windows.exe ./cmd/export-engine

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf dist/
	rm -rf node-wrapper/dist/
	rm -rf node-wrapper/bin/

# Run tests
test:
	go test -v ./...

# Run the server (example)
run: build
	./bin/$(BINARY)

# Development: build and copy to node-wrapper
dev: build-macos
	mkdir -p node-wrapper/bin
	cp bin/$(BINARY)-macos node-wrapper/bin/

# Node wrapper commands
install-node:
	cd node-wrapper && npm install

build-node:
	cd node-wrapper && npm run build

# Full build: Go + Node
build-full: build-all build-node
	mkdir -p node-wrapper/bin
	cp bin/$(BINARY)-macos node-wrapper/bin/
	cp bin/$(BINARY)-linux node-wrapper/bin/
	cp bin/$(BINARY)-windows.exe node-wrapper/bin/

# Quick test: build and test split-zip
test-split-zip: build
	./bin/$(BINARY) split-zip --input='[{"id":1,"name":"test"},{"id":2,"name":"demo"}]' \
		--output=/tmp/test.zip --format=csv --chunk-size=1 --split --zip

# Example servers
run-express: dev install-node
	cd example/express-export-test && npm install && node server.js

run-nestjs: dev install-node build-node
	cd example/nestjs-example && npm install && npm run start:dev

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build for current platform"
	@echo "  build-all     - Build for macOS, Linux, Windows"
	@echo "  build-full    - Build Go + Node wrapper"
	@echo "  clean         - Remove build artifacts"
	@echo "  test          - Run Go tests"
	@echo "  dev           - Build and copy to node-wrapper/bin"
	@echo "  install-node  - Install node-wrapper dependencies"
	@echo "  build-node    - Build node-wrapper TypeScript"
	@echo "  run-express   - Run Express example server"
	@echo "  run-nestjs    - Run NestJS example server"
	@echo "  test-split-zip- Quick test of split-zip feature"
	@echo "  help          - Show this help"
