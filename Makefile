.PHONY: build run test clean dev

# Build the application
build:
	go build -o bin/server backend/cmd/main.go

# Run the application
run:
	go run backend/cmd/main.go

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -cover ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Install dependencies
deps:
	go mod init task
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Run database migrations (manual)
migrate:
	go run cmd/server/main.go -migrate