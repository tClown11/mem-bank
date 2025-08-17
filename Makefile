.PHONY: build run test clean docker-build docker-run migrate-up migrate-down

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=mem_bank
BINARY_UNIX=$(BINARY_NAME)_unix

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/api

# Run the application
run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/api
	./$(BINARY_NAME)

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

# Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Build for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v ./cmd/api

# Docker commands
docker-build:
	docker build -t mem_bank:latest .

docker-run:
	docker-compose up --build

docker-down:
	docker-compose down

docker-clean:
	docker-compose down -v
	docker rmi mem_bank:latest

# Database migration commands (requires golang-migrate)
migrate-install:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate-up:
	migrate -path migrations -database "postgresql://mem_bank_user:mem_bank_password@localhost:5432/mem_bank?sslmode=disable" -verbose up

migrate-down:
	migrate -path migrations -database "postgresql://mem_bank_user:mem_bank_password@localhost:5432/mem_bank?sslmode=disable" -verbose down

migrate-force:
	migrate -path migrations -database "postgresql://mem_bank_user:mem_bank_password@localhost:5432/mem_bank?sslmode=disable" -verbose force $(VERSION)

# Development commands
dev-setup:
	docker-compose up postgres redis -d
	sleep 10
	make migrate-up

dev-run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/api
	JWT_SECRET=dev-secret ./$(BINARY_NAME)

# Linting
lint:
	golangci-lint run

# Format code
fmt:
	$(GOCMD) fmt ./...

# Vet code
vet:
	$(GOCMD) vet ./...