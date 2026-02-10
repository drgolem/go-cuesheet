# Build configuration
BINDIR := ./bin
TOOLS := normalize-cue decode-mojibake print-tracks

.PHONY: all build clean test lint tools help

# Default target
all: build

# Build all tools
build: tools

# Build individual tools into bin/
tools:
	@mkdir -p $(BINDIR)
	@echo "Building tools..."
	@go build -o $(BINDIR)/normalize-cue ./tools/normalize-cue
	@go build -o $(BINDIR)/decode-mojibake ./tools/decode-mojibake
	@go build -o $(BINDIR)/print-tracks ./examples/print-tracks
	@echo "✓ Tools built successfully in $(BINDIR)/"

# Build specific tool
normalize-cue:
	@mkdir -p $(BINDIR)
	@go build -o $(BINDIR)/normalize-cue ./tools/normalize-cue
	@echo "✓ Built normalize-cue"

decode-mojibake:
	@mkdir -p $(BINDIR)
	@go build -o $(BINDIR)/decode-mojibake ./tools/decode-mojibake
	@echo "✓ Built decode-mojibake"

print-tracks:
	@mkdir -p $(BINDIR)
	@go build -o $(BINDIR)/print-tracks ./examples/print-tracks
	@echo "✓ Built print-tracks"

# Run all tests
test:
	@echo "Running tests..."
	@go test -cover ./cuesheet
	@go test -cover ./cuesheet/encoding
	@echo "✓ All tests passed"

# Run tests with verbose output
test-verbose:
	@go test -v -cover ./cuesheet
	@go test -v -cover ./cuesheet/encoding

# Run linter
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install from https://golangci-lint.run/"; exit 1)
	@golangci-lint run ./...

# Clean build artifacts
clean:
	@rm -rf $(BINDIR)
	@echo "✓ Cleaned build artifacts"

# Install tools to GOPATH/bin
install: build
	@cp $(BINDIR)/* $(shell go env GOPATH)/bin/
	@echo "✓ Tools installed to $(shell go env GOPATH)/bin/"

# Show help
help:
	@echo "Available targets:"
	@echo "  make build          - Build all tools (default)"
	@echo "  make tools          - Build all tools"
	@echo "  make normalize-cue  - Build normalize-cue tool only"
	@echo "  make decode-mojibake - Build decode-mojibake tool only"
	@echo "  make print-tracks   - Build print-tracks example"
	@echo "  make test           - Run all tests with coverage"
	@echo "  make test-verbose   - Run tests with verbose output"
	@echo "  make lint           - Run linter"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make install        - Install tools to GOPATH/bin"
	@echo "  make help           - Show this help message"
