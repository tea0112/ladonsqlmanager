# 🎮 Ladon Policy Playground Makefile
# Make it easy to build and run playground tools

.PHONY: all build clean cli quick-test run-cli run-quick-test help

# Default target
all: build

# Build all tools
build: cli quick-test
	@echo "✅ All playground tools built successfully!"

# Build CLI tool
cli:
	@echo "🔨 Building policy CLI..."
	@go build -o policy_cli policy_cli.go
	@echo "✅ CLI tool built: policy_cli"

# Build quick test tool
quick-test:
	@echo "🔨 Building quick test tool..."
	@go build -o quick_test quick_test_main.go
	@echo "✅ Quick test tool built: quick_test"

# Run CLI tool
run-cli: cli
	@echo "🚀 Running Interactive Policy CLI..."
	@./policy_cli

# Run quick test tool
run-quick-test: quick-test
	@echo "🧪 Running Quick Test Scenarios..."
	@./quick_test

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -f policy_cli quick_test
	@echo "✅ Cleaned!"

# Show help
help:
	@echo "🎮 Ladon Policy Playground Makefile"
	@echo "=================================="
	@echo ""
	@echo "Available targets:"
	@echo "  all           - Build all tools (default)"
	@echo "  build         - Build all tools"
	@echo "  cli           - Build only the CLI tool"
	@echo "  quick-test    - Build only the quick test tool"
	@echo "  run-cli       - Build and run the CLI tool"
	@echo "  run-quick-test- Build and run the quick test tool"
	@echo "  clean         - Remove build artifacts"
	@echo "  help          - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make                    # Build all tools"
	@echo "  make run-cli           # Build and run CLI"
	@echo "  make run-quick-test    # Build and run quick test"
	@echo "  make clean             # Clean build files"
	@echo ""
	@echo "Prerequisites:"
	@echo "  - Set DB_STRING environment variable or create config.env"
	@echo "  - Database schema will be created automatically via GORM"
	@echo "  - Go 1.16+ installed"
	@echo ""
	@echo "Configuration:"
	@echo "  - Create config.env with DB_STRING for automatic configuration"
	@echo "  - Or set DB_STRING environment variable manually"
