.PHONY: build run test clean lint help

# Binary name
BINARY_NAME=bonpreu-go

# Build directory
BUILD_DIR=build

# Main application path
MAIN_PATH=cmd/bonpreu/main.go

# Default target
.DEFAULT_GOAL := help

# Help target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build the application
build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the application
run: ## Run the application
	@echo "Running $(BINARY_NAME)..."
	@go run $(MAIN_PATH)

# Test the application
test: ## Run tests
	@echo "Running tests..."
	@go test ./...

# Clean build artifacts
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@go clean

# Install dependencies
deps: ## Install dependencies
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download

# Lint the code
lint: ## Lint the code
	@echo "Linting code..."
	@go vet ./...
	@golangci-lint run

# Format code
fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

# Check for security vulnerabilities
security: ## Check for security vulnerabilities
	@echo "Checking for security vulnerabilities..."
	@govulncheck ./...

# Run with race detection
race: ## Run with race detection
	@echo "Running with race detection..."
	@go run -race $(MAIN_PATH)

# Generate documentation
docs: ## Generate documentation
	@echo "Generating documentation..."
	@go doc -all ./...

# Install the application
install: build ## Install the application
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installation complete"

# Uninstall the application
uninstall: ## Uninstall the application
	@echo "Uninstalling $(BINARY_NAME)..."
	@rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Uninstallation complete" 