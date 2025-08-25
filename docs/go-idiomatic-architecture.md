# Go-Idiomatic Architecture Design

## Overview

This document outlines a proper Go-idiomatic architecture for the Memory Bank system that adheres to Go design principles, follows the standard Go project layout, and enables proper separation of concerns.

## Proposed Architecture Principles

### 1. **Package-First Design**
- Each package has a single, well-defined purpose
- Dependencies flow in one direction (downward)
- No circular dependencies
- Interfaces defined where they're consumed

### 2. **Dependency Inversion via Interfaces**
- High-level modules define interfaces they need
- Low-level modules implement those interfaces
- Dependency injection through constructors, not service containers

### 3. **Clean Domain Separation**
- Domain entities contain only business logic
- No infrastructure concerns in domain layer
- Repository interfaces defined in domain, implementations in infrastructure

### 4. **Go-Standard Project Layout**
```
/mem_bank
├── cmd/                    # Main applications
├── internal/               # Private application code
│   ├── app/               # Application orchestration layer
│   ├── domain/            # Core business entities and interfaces
│   ├── usecase/           # Business logic implementations
│   ├── repository/        # Data access layer
│   └── transport/         # Delivery mechanisms (HTTP, gRPC, etc.)
├── pkg/                   # Public library code
└── api/                   # API definitions
```

## Detailed Architecture Design

### 1. **Domain Layer** (`internal/domain/`)

#### 1.1 Core Entities
```go
// internal/domain/user/entity.go
package user

import (
    "time"
    "github.com/google/uuid"
)

// User represents a core business entity
type User struct {
    ID        ID
    Username  string
    Email     string
    Profile   Profile
    Settings  Settings
    CreatedAt time.Time
    UpdatedAt time.Time
    LastLogin time.Time
    IsActive  bool
}

type ID uuid.UUID

type Profile struct {
    FirstName   string
    LastName    string
    Avatar      string
    Bio         string
    Preferences map[string]interface{}
}

type Settings struct {
    Language             string
    Timezone             string
    MemoryRetention      int
    PrivacyLevel         string
    NotificationSettings map[string]bool
    EmbeddingModel       string
    MaxMemories          int
    AutoSummary          bool
}
```

#### 1.2 Repository Interfaces (Defined in Domain)
```go
// internal/domain/user/repository.go
package user

import "context"

// Repository defines data access interface for users
type Repository interface {
    Store(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id ID) (*User, error)
    FindByUsername(ctx context.Context, username string) (*User, error)
    FindByEmail(ctx context.Context, email string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id ID) error
}
```

#### 1.3 Service Interfaces (Defined in Domain)
```go
// internal/domain/user/service.go
package user

import "context"

// Service defines business operations for users
type Service interface {
    CreateUser(ctx context.Context, req CreateRequest) (*User, error)
    GetUser(ctx context.Context, id ID) (*User, error)
    UpdateUser(ctx context.Context, id ID, req UpdateRequest) (*User, error)
    DeleteUser(ctx context.Context, id ID) error
    ListUsers(ctx context.Context, limit, offset int) ([]*User, error)
}
```

#### 1.4 Value Objects and Errors
```go
// internal/domain/user/errors.go
package user

import "errors"

var (
    ErrNotFound      = errors.New("user not found")
    ErrAlreadyExists = errors.New("user already exists")
    ErrInvalidEmail  = errors.New("invalid email format")
    ErrInvalidUsername = errors.New("invalid username")
)

// internal/domain/user/requests.go
package user

type CreateRequest struct {
    Username string
    Email    string
    Profile  Profile
    Settings Settings
}

type UpdateRequest struct {
    Username *string
    Email    *string
    Profile  *Profile
    Settings *Settings
    IsActive *bool
}
```

### 2. **Use Case Layer** (`internal/usecase/`)

```go
// internal/usecase/user/service.go
package user

import (
    "context"
    "fmt"
    
    "mem_bank/internal/domain/user"
)

// service implements user.Service interface
type service struct {
    repo user.Repository
}

// NewService creates a new user service
func NewService(repo user.Repository) user.Service {
    return &service{
        repo: repo,
    }
}

func (s *service) CreateUser(ctx context.Context, req user.CreateRequest) (*user.User, error) {
    // Validation
    if err := s.validateCreateRequest(req); err != nil {
        return nil, err
    }
    
    // Business logic
    existingUser, _ := s.repo.FindByUsername(ctx, req.Username)
    if existingUser != nil {
        return nil, user.ErrAlreadyExists
    }
    
    // Create entity
    u := &user.User{
        ID:       user.ID(uuid.New()),
        Username: req.Username,
        Email:    req.Email,
        Profile:  req.Profile,
        Settings: req.Settings,
        IsActive: true,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    
    // Apply defaults
    s.setDefaultSettings(&u.Settings)
    
    // Store
    if err := s.repo.Store(ctx, u); err != nil {
        return nil, fmt.Errorf("storing user: %w", err)
    }
    
    return u, nil
}

func (s *service) validateCreateRequest(req user.CreateRequest) error {
    // Validation logic here
    return nil
}
```

### 3. **Repository Layer** (`internal/repository/`)

```go
// internal/repository/user/postgres.go
package user

import (
    "context"
    "encoding/json"
    
    "gorm.io/gorm"
    
    "mem_bank/internal/domain/user"
)

// postgresRepository implements user.Repository
type postgresRepository struct {
    db *gorm.DB
}

// NewPostgresRepository creates a new postgres-based user repository
func NewPostgresRepository(db *gorm.DB) user.Repository {
    return &postgresRepository{db: db}
}

// userModel represents the database model
type userModel struct {
    ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    Username  string    `gorm:"uniqueIndex;not null"`
    Email     string    `gorm:"uniqueIndex;not null"`
    Profile   string    `gorm:"type:jsonb"`
    Settings  string    `gorm:"type:jsonb"`
    CreatedAt time.Time `gorm:"not null"`
    UpdatedAt time.Time `gorm:"not null"`
    LastLogin *time.Time
    IsActive  bool      `gorm:"default:true"`
}

func (r *postgresRepository) Store(ctx context.Context, user *user.User) error {
    model, err := r.toModel(user)
    if err != nil {
        return fmt.Errorf("converting to model: %w", err)
    }
    
    if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
        return fmt.Errorf("creating user: %w", err)
    }
    
    return nil
}

func (r *postgresRepository) FindByID(ctx context.Context, id user.ID) (*user.User, error) {
    var model userModel
    if err := r.db.WithContext(ctx).Where("id = ?", string(id)).First(&model).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, user.ErrNotFound
        }
        return nil, fmt.Errorf("finding user: %w", err)
    }
    
    return r.toDomain(&model)
}

func (r *postgresRepository) toModel(u *user.User) (*userModel, error) {
    profile, err := json.Marshal(u.Profile)
    if err != nil {
        return nil, err
    }
    
    settings, err := json.Marshal(u.Settings)
    if err != nil {
        return nil, err
    }
    
    return &userModel{
        ID:        string(u.ID),
        Username:  u.Username,
        Email:     u.Email,
        Profile:   string(profile),
        Settings:  string(settings),
        CreatedAt: u.CreatedAt,
        UpdatedAt: u.UpdatedAt,
        LastLogin: nilTimePtr(u.LastLogin),
        IsActive:  u.IsActive,
    }, nil
}

func (r *postgresRepository) toDomain(m *userModel) (*user.User, error) {
    var profile user.Profile
    if err := json.Unmarshal([]byte(m.Profile), &profile); err != nil {
        return nil, err
    }
    
    var settings user.Settings
    if err := json.Unmarshal([]byte(m.Settings), &settings); err != nil {
        return nil, err
    }
    
    return &user.User{
        ID:        user.ID(uuid.MustParse(m.ID)),
        Username:  m.Username,
        Email:     m.Email,
        Profile:   profile,
        Settings:  settings,
        CreatedAt: m.CreatedAt,
        UpdatedAt: m.UpdatedAt,
        LastLogin: derefTime(m.LastLogin),
        IsActive:  m.IsActive,
    }, nil
}
```

### 4. **Transport Layer** (`internal/transport/`)

```go
// internal/transport/http/user/handler.go
package user

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    
    "mem_bank/internal/domain/user"
)

// Handler handles HTTP requests for user operations
type Handler struct {
    service user.Service
}

// NewHandler creates a new user HTTP handler
func NewHandler(service user.Service) *Handler {
    return &Handler{
        service: service,
    }
}

func (h *Handler) CreateUser(c *gin.Context) {
    var req user.CreateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    user, err := h.service.CreateUser(c.Request.Context(), req)
    if err != nil {
        h.handleError(c, err)
        return
    }
    
    c.JSON(http.StatusCreated, h.toResponse(user))
}

func (h *Handler) handleError(c *gin.Context, err error) {
    switch err {
    case user.ErrNotFound:
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
    case user.ErrAlreadyExists:
        c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
    case user.ErrInvalidEmail, user.ErrInvalidUsername:
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    default:
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
    }
}

func (h *Handler) toResponse(u *user.User) interface{} {
    return map[string]interface{}{
        "id":         u.ID,
        "username":   u.Username,
        "email":      u.Email,
        "profile":    u.Profile,
        "settings":   u.Settings,
        "created_at": u.CreatedAt,
        "updated_at": u.UpdatedAt,
        "last_login": u.LastLogin,
        "is_active":  u.IsActive,
    }
}
```

### 5. **Application Layer** (`internal/app/`)

```go
// internal/app/app.go
package app

import (
    "context"
    "fmt"
    "net/http"
    
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    
    userUsecase "mem_bank/internal/usecase/user"
    userRepo "mem_bank/internal/repository/user"
    userHandler "mem_bank/internal/transport/http/user"
)

// App represents the application
type App struct {
    server *http.Server
    db     *gorm.DB
}

// NewApp creates a new application instance
func NewApp(db *gorm.DB) *App {
    return &App{
        db: db,
    }
}

// Start starts the application
func (a *App) Start(ctx context.Context, port string) error {
    // Wire up dependencies
    userRepository := userRepo.NewPostgresRepository(a.db)
    userService := userUsecase.NewService(userRepository)
    userHandler := userHandler.NewHandler(userService)
    
    // Setup router
    router := gin.New()
    a.setupRoutes(router, userHandler)
    
    // Setup server
    a.server = &http.Server{
        Addr:    ":" + port,
        Handler: router,
    }
    
    return a.server.ListenAndServe()
}

func (a *App) setupRoutes(router *gin.Engine, userHandler *userHandler.Handler) {
    api := router.Group("/api/v1")
    
    users := api.Group("/users")
    {
        users.POST("", userHandler.CreateUser)
        users.GET("/:id", userHandler.GetUser)
        users.PUT("/:id", userHandler.UpdateUser)
        users.DELETE("/:id", userHandler.DeleteUser)
        users.GET("", userHandler.ListUsers)
    }
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown(ctx context.Context) error {
    return a.server.Shutdown(ctx)
}
```

### 6. **Main Application** (`cmd/api/main.go`)

```go
// cmd/api/main.go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    
    "mem_bank/configs"
    "mem_bank/internal/app"
    "mem_bank/pkg/database"
    "mem_bank/pkg/logger"
)

func main() {
    // Load configuration
    config, err := configs.LoadConfig("")
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    // Initialize logger
    appLogger, err := logger.NewLogger(&config.Logging)
    if err != nil {
        log.Fatalf("Failed to initialize logger: %v", err)
    }
    
    // Initialize database
    db, err := database.NewGormConnection(&config.Database)
    if err != nil {
        appLogger.WithError(err).Fatal("Failed to connect to database")
    }
    defer db.Close()
    
    // Create and start application
    application := app.NewApp(db.DB)
    
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Start server in goroutine
    go func() {
        if err := application.Start(ctx, config.Server.Port); err != nil {
            appLogger.WithError(err).Fatal("Failed to start server")
        }
    }()
    
    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    // Graceful shutdown
    appLogger.Info("Shutting down server...")
    if err := application.Shutdown(ctx); err != nil {
        appLogger.WithError(err).Error("Server forced to shutdown")
    }
}
```

## Key Architectural Benefits

### 1. **Proper Separation of Concerns**
- Each layer has a single responsibility
- Clear boundaries between business logic, data access, and presentation
- Easy to understand and maintain

### 2. **Dependency Inversion**
- High-level modules define interfaces they need
- Low-level modules implement those interfaces
- Easy to test and mock

### 3. **Package Cohesion**
- Related functionality grouped in focused packages
- Clear package boundaries and responsibilities
- Follows Go package philosophy

### 4. **Testability**
- Each layer can be tested independently
- Easy to create mocks and test doubles
- Clear separation enables unit testing

### 5. **Extensibility**
- Easy to add new features without affecting existing code
- New implementations can be plugged in easily
- Supports multiple delivery mechanisms (HTTP, gRPC, CLI)

### 6. **Go Idioms Compliance**
- Accepts interfaces, returns structs
- Small, focused interfaces
- Constructor-based dependency injection
- Proper error handling patterns

This architecture follows Go best practices and provides a solid foundation for building maintainable, testable, and extensible applications.