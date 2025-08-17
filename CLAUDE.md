# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based AI memory bank system implementation based on the architectural blueprint in `docs/arch.md`. The project aims to build a production-grade AI memory layer using Golang, Qdrant, and pgvector.

## Common Commands

### Build and Run
```bash
go run main.go                  # Run the application
go build -o mem_bank main.go    # Build binary
```

### Go Module Management
```bash
go mod tidy                     # Clean up dependencies
go mod download                 # Download dependencies
go get <package>                # Add new dependency
```

### Testing
```bash
go test ./...                   # Run all tests
go test -v ./...                # Run tests with verbose output
go test -race ./...             # Run tests with race detection
```

### Code Quality
```bash
go fmt ./...                    # Format code
go vet ./...                    # Examine Go source code and reports suspicious constructs
```

## Architecture Overview

The project follows Clean Architecture principles with the following planned structure:

### Core Components
- **Domain Layer**: Core entities and interfaces (memory.go, user.go)
- **Use Case Layer**: Business logic implementation (memory_usecase.go)
- **Repository Layer**: Data access implementations (qdrant_repo.go, postgres_repo.go)
- **Delivery Layer**: HTTP handlers and routing

### Key Design Patterns
- **Strategy Pattern**: Pluggable database backends (Qdrant vs PostgreSQL/pgvector)
- **Repository Pattern**: Abstract data access layer
- **Dependency Injection**: Loose coupling between layers

### Two-Stage Memory Pipeline
1. **Extraction Stage**: LLM-powered extraction of candidate memories from raw input
2. **Update Stage**: Intelligent decision-making (ADD/UPDATE/DELETE/NOOP) using LLM tool calls

### Technology Stack
- **Language**: Go 1.24+
- **Web Framework**: Gin (planned)
- **Primary Vector DB**: Qdrant
- **Alternative Vector DB**: PostgreSQL with pgvector
- **Configuration**: Viper
- **Containerization**: Docker & Docker Compose

## Development Guidelines

### Project Structure (Planned)
```
/mem_bank
├── cmd/api/main.go              # Application entry point
├── configs/config.yaml          # Configuration files
├── internal/
│   ├── domain/                  # Core entities and interfaces
│   ├── usecase/                 # Business logic
│   ├── repository/              # Data access implementations
│   └── delivery/http/           # HTTP handlers and routing
├── pkg/llm/                     # LLM abstraction layer
└── docs/arch.md                 # Comprehensive architecture documentation
```

### Key Architectural Principles
- Use interfaces for dependency inversion
- Implement asynchronous processing for memory operations
- Support concurrent operations using Go's goroutines
- Maintain database-agnostic design through repository abstraction

### Configuration Management
- Use environment variables for deployment-specific settings
- Support multiple configuration sources (files, env vars, CLI flags)
- Follow 12-factor app principles

## Important Notes

- The main.go file is currently minimal (just package declaration)
- Comprehensive architecture documentation is available in `docs/arch.md` (Chinese)
- The project is designed to be a Memory Component Provider (MCP) for AI applications
- Asynchronous processing is critical for production performance
- The system supports pluggable vector database backends