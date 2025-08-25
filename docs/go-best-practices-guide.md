# Go Best Practices Guide

## Overview

This guide provides comprehensive Go best practices specifically tailored for the Memory Bank project. It covers architectural patterns, coding standards, and development practices that ensure maintainable, testable, and performant Go code.

## 1. Package Design Principles

### 1.1 Package Organization
```go
// ✅ Good: Packages organized by domain/functionality
internal/
├── domain/
│   ├── user/           # User domain logic
│   └── memory/         # Memory domain logic
├── usecase/
│   ├── user/           # User business logic
│   └── memory/         # Memory business logic
└── repository/
    ├── user/           # User data access
    └── memory/         # Memory data access

// ❌ Bad: Packages organized by technical layer
internal/
├── models/             # All entities mixed together
├── services/           # All business logic mixed
└── repositories/       # All data access mixed
```

### 1.2 Package Naming
```go
// ✅ Good: Clear, concise package names
package user     // Not "userservice" or "users"
package memory   // Not "memorysvc" or "memories"
package postgres // Not "postgresrepository"

// ❌ Bad: Verbose or unclear names
package userservice
package memorymanagement
package databaserepository
```

### 1.3 Package Cohesion
```go
// ✅ Good: Related functionality in same package
package user

type User struct { ... }
type Repository interface { ... }
type Service interface { ... }
var ErrNotFound = errors.New("user not found")

// ❌ Bad: Unrelated functionality mixed
package models

type User struct { ... }
type Memory struct { ... }
type DatabaseConnection struct { ... }
```

## 2. Interface Design

### 2.1 Interface Placement
```go
// ✅ Good: Interfaces defined where they are consumed
// internal/domain/user/service.go
package user

// Service is consumed by HTTP handlers and defined in domain
type Service interface {
    CreateUser(ctx context.Context, req CreateRequest) (*User, error)
    GetUser(ctx context.Context, id ID) (*User, error)
}

// ❌ Bad: Interfaces defined where they are implemented
// internal/usecase/user/interfaces.go
package user

type UserService interface { ... } // Wrong location
```

### 2.2 Interface Size
```go
// ✅ Good: Small, focused interfaces
type Reader interface {
    Read(ctx context.Context, id ID) (*User, error)
}

type Writer interface {
    Write(ctx context.Context, user *User) error
}

type Repository interface {
    Reader
    Writer
}

// ❌ Bad: Large, monolithic interfaces
type UserRepository interface {
    Create(user *User) error
    GetByID(id uuid.UUID) (*User, error)
    GetByUsername(username string) (*User, error)
    GetByEmail(email string) (*User, error)
    Update(user *User) error
    Delete(id uuid.UUID) error
    List(limit, offset int) ([]*User, error)
    Count() (int, error)
    UpdateLastLogin(id uuid.UUID) error
    UpdateSettings(id uuid.UUID, settings UserSettings) error
    UpdateProfile(id uuid.UUID, profile UserProfile) error
}
```

### 2.3 Interface Composition
```go
// ✅ Good: Compose interfaces for different needs
type UserReader interface {
    FindByID(ctx context.Context, id ID) (*User, error)
    FindByEmail(ctx context.Context, email string) (*User, error)
}

type UserWriter interface {
    Store(ctx context.Context, user *User) error
    Delete(ctx context.Context, id ID) error
}

// Full repository combines both
type UserRepository interface {
    UserReader
    UserWriter
}

// Services that only need read access can depend on UserReader
func NewStatisticsService(reader UserReader) *StatisticsService {
    return &StatisticsService{reader: reader}
}
```

## 3. Dependency Injection

### 3.1 Constructor-Based Injection
```go
// ✅ Good: Clear constructor with interface dependencies
type userService struct {
    repo      user.Repository
    validator user.Validator
    logger    logger.Logger
}

func NewUserService(repo user.Repository, validator user.Validator, logger logger.Logger) user.Service {
    return &userService{
        repo:      repo,
        validator: validator,
        logger:    logger,
    }
}

// ❌ Bad: Service locator pattern
type userService struct {
    container *ServiceContainer
}

func (s *userService) CreateUser(req CreateRequest) (*User, error) {
    repo := s.container.GetUserRepository() // Bad: Hidden dependency
    validator := s.container.GetValidator()
    // ...
}
```

### 3.2 Avoiding Service Containers
```go
// ✅ Good: Explicit dependency wiring
func wireUserFeature(db *gorm.DB) (*userhttp.Handler, error) {
    repo := userrepo.NewPostgresRepository(db)
    service := userusecase.NewService(repo)
    handler := userhttp.NewHandler(service)
    return handler, nil
}

// ❌ Bad: Service container with mixed concerns
type Services struct {
    UserRepo      domain.UserRepository
    MemoryRepo    domain.MemoryRepository
    UserUsecase   usecase.UserUsecase
    MemoryUsecase usecase.MemoryUsecase
    UserHandler   *handlers.UserHandler
    MemoryHandler *handlers.MemoryHandler
}
```

## 4. Error Handling

### 4.1 Domain Errors
```go
// ✅ Good: Well-defined domain errors
package user

import "errors"

var (
    ErrNotFound      = errors.New("user not found")
    ErrAlreadyExists = errors.New("user already exists")
    ErrInvalidEmail  = errors.New("invalid email format")
)

// Custom error types for rich error context
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation error on field %s: %s", e.Field, e.Message)
}
```

### 4.2 Error Wrapping
```go
// ✅ Good: Proper error wrapping with context
func (s *userService) CreateUser(ctx context.Context, req CreateRequest) (*User, error) {
    if err := s.validator.Validate(req); err != nil {
        return nil, fmt.Errorf("validating create request: %w", err)
    }
    
    user := s.buildUser(req)
    if err := s.repo.Store(ctx, user); err != nil {
        return nil, fmt.Errorf("storing user %s: %w", user.Username, err)
    }
    
    return user, nil
}

// ❌ Bad: Error swallowing or poor context
func (s *userService) CreateUser(req CreateRequest) (*User, error) {
    err := s.validator.Validate(req)
    if err != nil {
        return nil, errors.New("validation failed") // Lost original error
    }
    
    user := s.buildUser(req)
    err = s.repo.Store(user)
    if err != nil {
        log.Error("Failed to store user") // Should return error, not just log
        return nil, err
    }
    
    return user, nil
}
```

### 4.3 Error Handling in HTTP Layer
```go
// ✅ Good: Proper error mapping in handlers
func (h *UserHandler) CreateUser(c *gin.Context) {
    var req user.CreateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    user, err := h.service.CreateUser(c.Request.Context(), req)
    if err != nil {
        h.handleServiceError(c, err)
        return
    }
    
    c.JSON(http.StatusCreated, h.toResponse(user))
}

func (h *UserHandler) handleServiceError(c *gin.Context, err error) {
    switch {
    case errors.Is(err, user.ErrNotFound):
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
    case errors.Is(err, user.ErrAlreadyExists):
        c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
    case errors.As(err, &user.ValidationError{}):
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    default:
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
    }
}
```

## 5. Testing Best Practices

### 5.1 Table-Driven Tests
```go
// ✅ Good: Comprehensive table-driven tests
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name    string
        req     user.CreateRequest
        setup   func(*usertest.MockRepository)
        want    *user.User
        wantErr error
    }{
        {
            name: "successful creation",
            req: user.CreateRequest{
                Username: "testuser",
                Email:    "test@example.com",
            },
            setup: func(repo *usertest.MockRepository) {
                repo.On("FindByUsername", mock.Anything, "testuser").Return(nil, user.ErrNotFound)
                repo.On("Store", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
            },
            want: &user.User{
                Username: "testuser",
                Email:    "test@example.com",
                IsActive: true,
            },
        },
        {
            name: "duplicate username",
            req: user.CreateRequest{
                Username: "existing",
                Email:    "new@example.com",
            },
            setup: func(repo *usertest.MockRepository) {
                existingUser := &user.User{Username: "existing"}
                repo.On("FindByUsername", mock.Anything, "existing").Return(existingUser, nil)
            },
            wantErr: user.ErrAlreadyExists,
        },
        {
            name: "invalid email",
            req: user.CreateRequest{
                Username: "testuser",
                Email:    "invalid-email",
            },
            setup:   func(repo *usertest.MockRepository) {},
            wantErr: user.ErrInvalidEmail,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            repo := &usertest.MockRepository{}
            tt.setup(repo)
            
            service := usecase.NewService(repo)
            got, err := service.CreateUser(context.Background(), tt.req)
            
            if tt.wantErr != nil {
                assert.Error(t, err)
                assert.ErrorIs(t, err, tt.wantErr)
                return
            }
            
            assert.NoError(t, err)
            assert.Equal(t, tt.want.Username, got.Username)
            assert.Equal(t, tt.want.Email, got.Email)
            assert.Equal(t, tt.want.IsActive, got.IsActive)
            assert.NotZero(t, got.ID)
            assert.NotZero(t, got.CreatedAt)
        })
    }
}
```

### 5.2 Test Organization
```go
// ✅ Good: Separate test package
package user_test

import (
    "testing"
    
    "mem_bank/internal/domain/user"
    "mem_bank/internal/usecase/user"
    "mem_bank/tests/testutil"
)

// ❌ Bad: Tests in same package
package user

import "testing"
```

### 5.3 Integration Tests
```go
// ✅ Good: Proper integration test setup
func TestUserRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    db := testutil.SetupTestDB(t)
    repo := NewPostgresRepository(db)
    
    t.Run("store and retrieve user", func(t *testing.T) {
        user := &user.User{
            ID:       user.ID(uuid.New()),
            Username: "testuser",
            Email:    "test@example.com",
            IsActive: true,
        }
        
        err := repo.Store(context.Background(), user)
        require.NoError(t, err)
        
        found, err := repo.FindByID(context.Background(), user.ID)
        require.NoError(t, err)
        assert.Equal(t, user.Username, found.Username)
        assert.Equal(t, user.Email, found.Email)
    })
}
```

## 6. Concurrency Patterns

### 6.1 Context Usage
```go
// ✅ Good: Proper context usage
func (s *userService) CreateUser(ctx context.Context, req CreateRequest) (*User, error) {
    // Check context cancellation
    if err := ctx.Err(); err != nil {
        return nil, err
    }
    
    // Pass context to repository
    if err := s.repo.Store(ctx, user); err != nil {
        return nil, fmt.Errorf("storing user: %w", err)
    }
    
    return user, nil
}

// Repository implementation respects context
func (r *postgresRepository) Store(ctx context.Context, user *User) error {
    return r.db.WithContext(ctx).Create(r.toModel(user)).Error
}
```

### 6.2 Goroutine Management
```go
// ✅ Good: Proper goroutine lifecycle management
func (s *notificationService) SendWelcomeEmail(ctx context.Context, user *User) error {
    // Use errgroup for managed concurrency
    var g errgroup.Group
    
    g.Go(func() error {
        return s.emailService.SendWelcome(ctx, user)
    })
    
    g.Go(func() error {
        return s.analyticsService.TrackSignup(ctx, user)
    })
    
    return g.Wait()
}

// ❌ Bad: Unmanaged goroutines
func (s *notificationService) SendWelcomeEmail(user *User) {
    go s.emailService.SendWelcome(user)     // No error handling
    go s.analyticsService.TrackSignup(user) // No context cancellation
}
```

## 7. Performance Best Practices

### 7.1 Memory Management
```go
// ✅ Good: Efficient memory usage
func (s *userService) ListUsers(ctx context.Context, limit, offset int) ([]*User, error) {
    // Pre-allocate slice with known capacity
    users := make([]*User, 0, limit)
    
    return s.repo.Find(ctx, limit, offset)
}

// Use sync.Pool for frequently allocated objects
var userPool = sync.Pool{
    New: func() interface{} {
        return &User{}
    },
}

func (s *userService) getUserFromPool() *User {
    return userPool.Get().(*User)
}

func (s *userService) returnUserToPool(user *User) {
    // Reset user fields
    *user = User{}
    userPool.Put(user)
}
```

### 7.2 Database Performance
```go
// ✅ Good: Efficient database queries
func (r *postgresRepository) FindUsersByIDs(ctx context.Context, ids []ID) ([]*User, error) {
    // Use bulk query instead of N+1 queries
    var models []userModel
    err := r.db.WithContext(ctx).
        Where("id IN ?", ids).
        Find(&models).Error
    if err != nil {
        return nil, err
    }
    
    users := make([]*User, len(models))
    for i, model := range models {
        users[i] = r.toDomain(&model)
    }
    
    return users, nil
}

// ❌ Bad: N+1 query pattern
func (r *postgresRepository) FindUsersByIDs(ctx context.Context, ids []ID) ([]*User, error) {
    var users []*User
    for _, id := range ids {
        user, err := r.FindByID(ctx, id) // Separate query for each ID
        if err != nil {
            return nil, err
        }
        users = append(users, user)
    }
    return users, nil
}
```

## 8. Configuration and Environment

### 8.1 Configuration Structure
```go
// ✅ Good: Structured configuration
type Config struct {
    Database DatabaseConfig `mapstructure:"database"`
    Server   ServerConfig   `mapstructure:"server"`
    Logger   LoggerConfig   `mapstructure:"logger"`
}

type DatabaseConfig struct {
    Host         string        `mapstructure:"host" env:"DB_HOST"`
    Port         int           `mapstructure:"port" env:"DB_PORT"`
    Username     string        `mapstructure:"username" env:"DB_USERNAME"`
    Password     string        `mapstructure:"password" env:"DB_PASSWORD"`
    Database     string        `mapstructure:"database" env:"DB_DATABASE"`
    MaxOpenConns int           `mapstructure:"max_open_conns" env:"DB_MAX_OPEN_CONNS"`
    MaxIdleConns int           `mapstructure:"max_idle_conns" env:"DB_MAX_IDLE_CONNS"`
    ConnMaxLife  time.Duration `mapstructure:"conn_max_life" env:"DB_CONN_MAX_LIFE"`
}
```

### 8.2 Environment-Based Configuration
```go
// ✅ Good: Environment-aware configuration loading
func LoadConfig(path string) (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    
    if path != "" {
        viper.AddConfigPath(path)
    }
    viper.AddConfigPath("./configs")
    viper.AddConfigPath(".")
    
    // Enable environment variable binding
    viper.AutomaticEnv()
    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    
    // Set defaults
    viper.SetDefault("server.port", 8080)
    viper.SetDefault("server.mode", "release")
    viper.SetDefault("database.max_open_conns", 10)
    
    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, fmt.Errorf("reading config file: %w", err)
        }
    }
    
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("unmarshaling config: %w", err)
    }
    
    return &config, nil
}
```

## 9. Logging and Observability

### 9.1 Structured Logging
```go
// ✅ Good: Structured logging with context
func (s *userService) CreateUser(ctx context.Context, req CreateRequest) (*User, error) {
    logger := s.logger.WithFields(logrus.Fields{
        "operation": "create_user",
        "username":  req.Username,
        "email":     req.Email,
    })
    
    logger.Info("creating user")
    
    user, err := s.buildUser(req)
    if err != nil {
        logger.WithError(err).Error("failed to build user")
        return nil, err
    }
    
    if err := s.repo.Store(ctx, user); err != nil {
        logger.WithError(err).Error("failed to store user")
        return nil, fmt.Errorf("storing user: %w", err)
    }
    
    logger.WithField("user_id", user.ID).Info("user created successfully")
    return user, nil
}
```

### 9.2 Metrics and Tracing
```go
// ✅ Good: Metrics collection
type instrumentedUserService struct {
    service user.Service
    metrics *prometheus.CounterVec
}

func (s *instrumentedUserService) CreateUser(ctx context.Context, req user.CreateRequest) (*user.User, error) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        s.metrics.WithLabelValues("create_user").Add(duration.Seconds())
    }()
    
    return s.service.CreateUser(ctx, req)
}
```

## 10. Security Best Practices

### 10.1 Input Validation
```go
// ✅ Good: Comprehensive input validation
type UserValidator struct {
    emailRegex    *regexp.Regexp
    usernameRegex *regexp.Regexp
}

func NewUserValidator() *UserValidator {
    return &UserValidator{
        emailRegex:    regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
        usernameRegex: regexp.MustCompile(`^[a-zA-Z0-9_]{3,50}$`),
    }
}

func (v *UserValidator) ValidateCreateRequest(req CreateRequest) error {
    var errs []string
    
    if !v.usernameRegex.MatchString(req.Username) {
        errs = append(errs, "username must be 3-50 characters and contain only letters, numbers, and underscores")
    }
    
    if !v.emailRegex.MatchString(req.Email) {
        errs = append(errs, "invalid email format")
    }
    
    if len(errs) > 0 {
        return &ValidationError{Errors: errs}
    }
    
    return nil
}
```

This guide provides a solid foundation for writing idiomatic, maintainable, and performant Go code in the Memory Bank project. Following these practices will result in a codebase that is easier to understand, test, and extend.