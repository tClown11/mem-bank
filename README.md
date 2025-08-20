# AI Memory Bank System

A production-grade AI memory layer implementation built with Go, designed to provide persistent memory capabilities for AI agents and applications. The system enables AI applications to maintain context, learn from interactions, and provide personalized experiences through intelligent memory management.

## 🎯 Project Overview

The AI Memory Bank System is a Memory Component Provider (MCP) that serves as the "long-term brain" for AI applications. It transforms stateless AI interactions into stateful, context-aware experiences by intelligently storing, retrieving, and managing memories from user interactions.

### Key Features

- **🧠 Intelligent Memory Pipeline**: LLM-powered two-stage memory processing (extraction and update)
- **🗄️ Dual Storage Backend**: Support for both PostgreSQL with pgvector and Qdrant (planned)
- **⚡ Asynchronous Processing**: Non-blocking memory operations for real-time AI interactions
- **🔍 Semantic Search**: Advanced vector similarity search with configurable thresholds
- **🏷️ Rich Metadata**: Support for tags, importance levels, memory types, and custom metadata
- **📊 Analytics & Statistics**: Comprehensive memory usage tracking and user statistics
- **🔒 Security**: JWT-based authentication with configurable rate limiting
- **🐳 Containerized**: Docker and Docker Compose support for easy deployment
- **🏗️ Clean Architecture**: Modular design following dependency injection principles

## 🏛️ Architecture Overview

The system follows Clean Architecture principles with clear separation of concerns:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Delivery      │    │    UseCase      │    │    Domain       │
│   (HTTP API)    │◄───│  (Business      │◄───│   (Entities &   │
│                 │    │   Logic)        │    │   Interfaces)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 ▼
                    ┌─────────────────┐
                    │   Repository    │
                    │ (Data Access)   │
                    └─────────────────┘
                             │
                    ┌─────────────────┐
                    │   PostgreSQL    │
                    │   + pgvector    │
                    └─────────────────┘
```

### Core Components

- **Domain Layer**: Core entities (`Memory`, `User`) and interfaces
- **Use Case Layer**: Business logic implementation (`MemoryUsecase`, `UserUsecase`)
- **Repository Layer**: Data access implementations with pluggable backends
- **Delivery Layer**: HTTP handlers and routing (Gin framework)
- **Infrastructure**: Database connections, logging, configuration management

## 🚀 Technology Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Language** | Go 1.24+ | High-performance backend with excellent concurrency |
| **Web Framework** | Gin | Fast HTTP router and middleware support |
| **Primary Database** | PostgreSQL + pgvector | Vector storage with relational data capabilities |
| **Alternative Database** | Qdrant (planned) | Specialized vector database for high-performance scenarios |
| **Caching** | Redis | Fast in-memory caching for frequently accessed data |
| **Configuration** | Viper | Multi-source configuration management |
| **Database Migrations** | golang-migrate | Version-controlled schema management |
| **Containerization** | Docker & Docker Compose | Consistent deployment environments |
| **ORM** | GORM | Database abstraction and code generation |
| **Logging** | Logrus | Structured logging with multiple output formats |

## 📦 Installation and Setup

### Prerequisites

- Go 1.24 or higher
- Docker and Docker Compose
- PostgreSQL (if running locally)
- Redis (if running locally)

### Quick Start with Docker

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd mem_bank
   ```

2. **Start the services**:
   ```bash
   docker-compose up --build
   ```

3. **The API will be available at**: `http://localhost:8080`

### Local Development Setup

1. **Install dependencies**:
   ```bash
   make deps
   ```

2. **Start infrastructure services**:
   ```bash
   make dev-setup
   ```

3. **Run database migrations**:
   ```bash
   make migrate-up
   ```

4. **Start the application**:
   ```bash
   make dev-run
   ```

### Configuration

The system uses a flexible configuration system supporting multiple sources:

- **Configuration file**: `configs/config.yaml`
- **Environment variables**: Override any config value
- **Command-line flags**: Highest priority

Key configuration sections:
- **Server**: HTTP server settings (port, timeouts)
- **Database**: PostgreSQL connection parameters
- **Redis**: Cache configuration
- **AI**: Embedding model settings and thresholds
- **Security**: JWT settings and rate limiting
- **Logging**: Log level and output format

## 🔧 API Reference

### Health Check

```http
GET /api/v1/health
```

### User Management

```http
POST   /api/v1/users           # Create user
GET    /api/v1/users/:id       # Get user by ID
PUT    /api/v1/users/:id       # Update user
DELETE /api/v1/users/:id       # Delete user
GET    /api/v1/users/stats     # Get user statistics
```

### Memory Operations

```http
POST   /api/v1/memories                    # Create memory
GET    /api/v1/memories/:id                # Get memory by ID
PUT    /api/v1/memories/:id                # Update memory
DELETE /api/v1/memories/:id                # Delete memory
GET    /api/v1/memories/user/:user_id      # Get user's memories
POST   /api/v1/memories/search/similar     # Semantic similarity search
GET    /api/v1/memories/search/content     # Text-based search
POST   /api/v1/memories/search/tags        # Tag-based search
```

### Request Examples

**Create Memory**:
```json
POST /api/v1/memories
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "content": "User prefers dark mode in applications",
  "summary": "UI preference for dark theme",
  "importance": 7,
  "memory_type": "preference",
  "tags": ["ui", "preference", "theme"],
  "metadata": {
    "category": "user_preference",
    "source": "settings_page"
  }
}
```

**Similarity Search**:
```json
POST /api/v1/memories/search/similar
{
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "query": "What are the user's UI preferences?",
  "limit": 10,
  "threshold": 0.8
}
```

## 📊 Database Schema

### Users Table
- `id` (UUID): Primary key
- `username` (VARCHAR): Unique username
- `email` (VARCHAR): Unique email address
- `profile` (JSONB): User profile information
- `settings` (JSONB): User-specific settings
- `created_at`, `updated_at`, `last_login` (TIMESTAMP)
- `is_active` (BOOLEAN): Account status

### Memories Table
- `id` (UUID): Primary key
- `user_id` (UUID): Foreign key to users table
- `content` (TEXT): The actual memory content
- `summary` (TEXT): Optional condensed summary
- `embedding` (VECTOR): 1536-dimensional embedding vector
- `importance` (INTEGER): 1-10 scale for memory importance
- `memory_type` (VARCHAR): Category of memory
- `tags` (TEXT[]): Searchable tags
- `metadata` (JSONB): Additional structured data
- `access_count` (INTEGER): Usage tracking
- `created_at`, `updated_at`, `last_accessed` (TIMESTAMP)

## 🧪 Testing

### Run Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test package
go test -v ./internal/usecase/...
```

### Test Structure

- **Unit Tests**: `tests/unit/` - Test individual components in isolation
- **Integration Tests**: `tests/integration/` - Test component interactions
- **Mock Objects**: `tests/mocks/` - Generated mock implementations
- **Test Utilities**: `tests/testutil/` - Shared testing infrastructure

## 🚀 Deployment

### Production Deployment

1. **Build the application**:
   ```bash
   make build-linux
   ```

2. **Build Docker image**:
   ```bash
   make docker-build
   ```

3. **Deploy with Docker Compose**:
   ```bash
   docker-compose -f docker-compose.prod.yml up -d
   ```

### Environment Variables

Key environment variables for production:

```env
# Database
DATABASE_HOST=your-db-host
DATABASE_PASSWORD=secure-password

# Security
JWT_SECRET=your-jwt-secret

# AI Services
OPENAI_API_KEY=your-openai-key

# Redis
REDIS_URL=redis://your-redis-host:6379
```

## 🔄 Development Workflow

### Making Changes

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make changes and test**:
   ```bash
   make test
   make lint
   ```

3. **Format code**:
   ```bash
   make fmt
   make vet
   ```

4. **Database changes**:
   ```bash
   # Create migration
   migrate create -ext sql -dir migrations -seq your_migration_name
   
   # Apply migration
   make migrate-up
   ```

### Code Generation

The project uses GORM Gen for database model generation:

```bash
# Regenerate models
go run cmd/gen/main.go
```

## 📈 Performance Considerations

- **Vector Search**: Uses pgvector with IVFFlat indexing for efficient similarity search
- **Caching**: Redis caching layer for frequently accessed memories
- **Connection Pooling**: Optimized database connection management
- **Asynchronous Processing**: Background job processing for heavy operations
- **Indexing**: Comprehensive database indexing strategy

## 🛡️ Security Features

- **JWT Authentication**: Secure token-based authentication
- **Rate Limiting**: Configurable API rate limiting
- **CORS Support**: Cross-origin resource sharing configuration
- **Input Validation**: Request validation and sanitization
- **SQL Injection Protection**: Parameterized queries and ORM usage

## 🤝 Contributing

We welcome contributions! Please see our contribution guidelines:

1. **Fork the repository**
2. **Create a feature branch**
3. **Make your changes**
4. **Add tests for new functionality**
5. **Ensure all tests pass**
6. **Submit a pull request**

### Development Commands

```bash
# Setup development environment
make dev-setup

# Run in development mode
make dev-run

# Run tests with coverage
make test-coverage

# Lint code
make lint

# Format code
make fmt
```

## 📋 Project Status

**Current Stage**: MVP (Minimum Viable Product)

### ✅ Completed Features

- Core memory CRUD operations
- User management system
- PostgreSQL + pgvector integration
- RESTful API with comprehensive endpoints
- Docker containerization
- Database migrations
- Unit and integration testing framework
- Configuration management
- Logging and middleware

### 🚧 In Development

- LLM-powered memory extraction pipeline
- Intelligent memory update decisions
- Asynchronous job processing
- Advanced semantic search capabilities
- Memory importance scoring
- Automatic summarization

### 📋 Roadmap

- **Phase 1**: Core functionality and PostgreSQL backend ✅
- **Phase 2**: LLM integration and intelligent processing 🚧
- **Phase 3**: Qdrant backend support and performance optimization
- **Phase 4**: Advanced analytics and memory decay mechanisms
- **Phase 5**: Multi-modal memory support (images, audio)

## 📞 Support

For questions, issues, or contributions:

- **Issues**: Open an issue on GitHub
- **Documentation**: Check the `/docs` directory for detailed documentation
- **Architecture**: See `/docs/arch.md` for comprehensive architecture details

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

---

**Built with ❤️ using Go and modern AI technologies**