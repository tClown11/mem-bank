# Migration Strategy: From Current Architecture to Go-Idiomatic Design

## Overview

This document outlines a step-by-step migration strategy to transform the current architecture into a proper Go-idiomatic design. The migration is designed to be incremental, allowing the system to remain functional throughout the transition.

## Migration Phases

### Phase 1: Preparation and Foundation (Immediate)
**Duration:** 1-2 days  
**Risk:** Low

#### 1.1 Create New Package Structure
```bash
# Create new directory structure
mkdir -p internal/domain/{user,memory}
mkdir -p internal/usecase/{user,memory}
mkdir -p internal/repository/{user,memory}
mkdir -p internal/transport/http/{user,memory}
mkdir -p internal/app
```

#### 1.2 Clean Domain Layer
Create pure domain entities without infrastructure concerns:

**Target Files to Create:**
- `internal/domain/user/entity.go`
- `internal/domain/user/repository.go`
- `internal/domain/user/service.go`
- `internal/domain/user/errors.go`
- `internal/domain/user/requests.go`

**Migration Steps:**
1. Extract pure business entities from current `internal/domain/user.go`
2. Remove JSON and DB tags from domain entities
3. Move repository interfaces to domain packages
4. Create value objects and business rules

#### 1.3 Create Migration Compatibility Layer
Create adapter interfaces to maintain backward compatibility during migration:

```go
// internal/migration/adapter.go
package migration

import (
    "mem_bank/internal"
    newuser "mem_bank/internal/domain/user"
    olduser "mem_bank/internal/domain"
)

// UserServiceAdapter adapts new service to old interface
type UserServiceAdapter struct {
    newService newuser.Service
}

func NewUserServiceAdapter(service newuser.Service) *UserServiceAdapter {
    return &UserServiceAdapter{newService: service}
}

// Implement old interface methods by delegating to new service
func (a *UserServiceAdapter) CreateUser(req *olduser.UserCreateRequest) (*olduser.User, error) {
    // Convert request and delegate to new service
    newReq := convertCreateRequest(req)
    newUser, err := a.newService.CreateUser(context.Background(), newReq)
    if err != nil {
        return nil, err
    }
    return convertUser(newUser), nil
}
```

### Phase 2: Domain and Use Case Migration (Week 1)
**Duration:** 3-5 days  
**Risk:** Medium

#### 2.1 Migrate User Domain
1. Implement new user domain package with clean entities
2. Create new user service implementation
3. Update existing code to use adapters

#### 2.2 Migrate User Use Cases
1. Create new user use case implementation
2. Implement proper dependency injection
3. Add comprehensive unit tests

#### 2.3 Testing Strategy
- Write comprehensive tests for new components
- Maintain existing integration tests
- Create test adapters for compatibility

**Example Test Structure:**
```go
// internal/usecase/user/service_test.go
package user_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    
    "mem_bank/internal/domain/user"
    usecase "mem_bank/internal/usecase/user"
)

func TestService_CreateUser(t *testing.T) {
    tests := []struct {
        name    string
        req     user.CreateRequest
        setup   func(*MockRepository)
        want    *user.User
        wantErr error
    }{
        {
            name: "successful creation",
            req: user.CreateRequest{
                Username: "testuser",
                Email:    "test@example.com",
            },
            setup: func(repo *MockRepository) {
                repo.On("FindByUsername", mock.Anything, "testuser").Return(nil, user.ErrNotFound)
                repo.On("Store", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
            },
            want: &user.User{
                Username: "testuser",
                Email:    "test@example.com",
            },
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            repo := &MockRepository{}
            tt.setup(repo)
            
            service := usecase.NewService(repo)
            got, err := service.CreateUser(context.Background(), tt.req)
            
            if tt.wantErr != nil {
                assert.Error(t, err)
                assert.ErrorIs(t, err, tt.wantErr)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want.Username, got.Username)
                assert.Equal(t, tt.want.Email, got.Email)
            }
        })
    }
}
```

### Phase 3: Repository Migration (Week 2)
**Duration:** 2-3 days  
**Risk:** High (Database Layer)

#### 3.1 Create New Repository Implementation
1. Implement repository interface in new package structure
2. Create database model conversions
3. Maintain existing database schema

#### 3.2 Repository Testing
```go
// internal/repository/user/postgres_test.go
package user_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "gorm.io/gorm"
    
    "mem_bank/internal/domain/user"
    repo "mem_bank/internal/repository/user"
    "mem_bank/tests/testutil"
)

func TestPostgresRepository_Store(t *testing.T) {
    db := testutil.SetupTestDB(t)
    repository := repo.NewPostgresRepository(db)
    
    u := &user.User{
        ID:       user.ID(uuid.New()),
        Username: "testuser",
        Email:    "test@example.com",
        IsActive: true,
    }
    
    err := repository.Store(context.Background(), u)
    assert.NoError(t, err)
    
    // Verify storage
    found, err := repository.FindByID(context.Background(), u.ID)
    assert.NoError(t, err)
    assert.Equal(t, u.Username, found.Username)
    assert.Equal(t, u.Email, found.Email)
}
```

#### 3.3 Database Migration Safety
- Run both old and new implementations in parallel
- Compare results to ensure consistency
- Rollback plan if issues arise

### Phase 4: Transport Layer Migration (Week 2-3)
**Duration:** 2-3 days  
**Risk:** Medium

#### 4.1 Create New HTTP Handlers
1. Implement handlers in new package structure
2. Create proper error handling
3. Add request/response validation

#### 4.2 API Compatibility
- Maintain existing API endpoints
- Ensure response format compatibility
- Create API versioning if needed

#### 4.3 Handler Testing
```go
// internal/transport/http/user/handler_test.go
package user_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    
    "mem_bank/internal/domain/user"
    handler "mem_bank/internal/transport/http/user"
)

func TestHandler_CreateUser(t *testing.T) {
    gin.SetMode(gin.TestMode)
    
    service := &MockService{}
    h := handler.NewHandler(service)
    
    req := user.CreateRequest{
        Username: "testuser",
        Email:    "test@example.com",
    }
    
    expectedUser := &user.User{
        ID:       user.ID(uuid.New()),
        Username: "testuser",
        Email:    "test@example.com",
    }
    
    service.On("CreateUser", mock.Anything, req).Return(expectedUser, nil)
    
    jsonReq, _ := json.Marshal(req)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("POST", "/users", bytes.NewBuffer(jsonReq))
    c.Request.Header.Set("Content-Type", "application/json")
    
    h.CreateUser(c)
    
    assert.Equal(t, http.StatusCreated, w.Code)
    service.AssertExpectations(t)
}
```

### Phase 5: Application Wiring (Week 3)
**Duration:** 1-2 days  
**Risk:** Medium

#### 5.1 Create Application Layer
1. Implement dependency injection in app layer
2. Remove service container dependency
3. Update main.go to use new app structure

#### 5.2 Configuration Migration
```go
// internal/app/config.go
package app

type Config struct {
    Database DatabaseConfig
    Server   ServerConfig
    Logging  LoggingConfig
}

type Dependencies struct {
    UserService    user.Service
    MemoryService  memory.Service
    UserHandler    *userhttp.Handler
    MemoryHandler  *memoryhttp.Handler
}

func Wire(config Config) (*Dependencies, error) {
    // Database setup
    db, err := setupDatabase(config.Database)
    if err != nil {
        return nil, err
    }
    
    // Repositories
    userRepo := userrepo.NewPostgresRepository(db)
    memoryRepo := memoryrepo.NewPostgresRepository(db)
    
    // Services
    userService := userusecase.NewService(userRepo)
    memoryService := memoryusecase.NewService(memoryRepo, userRepo)
    
    // Handlers
    userHandler := userhttp.NewHandler(userService)
    memoryHandler := memoryhttp.NewHandler(memoryService)
    
    return &Dependencies{
        UserService:    userService,
        MemoryService:  memoryService,
        UserHandler:    userHandler,
        MemoryHandler:  memoryHandler,
    }, nil
}
```

### Phase 6: Cleanup and Optimization (Week 4)
**Duration:** 2-3 days  
**Risk:** Low

#### 6.1 Remove Old Code
1. Remove `internal/services.go`
2. Delete old package implementations
3. Clean up unused imports and dependencies

#### 6.2 Performance Testing
1. Run performance benchmarks
2. Compare with previous implementation
3. Optimize if necessary

#### 6.3 Documentation Update
1. Update code documentation
2. Create architectural decision records
3. Update deployment documentation

## Migration Checklist

### Pre-Migration
- [ ] Backup current codebase
- [ ] Create comprehensive test suite for current functionality
- [ ] Set up staging environment for testing
- [ ] Plan rollback strategy

### During Migration
- [ ] Run tests continuously
- [ ] Monitor application performance
- [ ] Maintain API compatibility
- [ ] Document any issues encountered

### Post-Migration
- [ ] Performance validation
- [ ] Security review
- [ ] Code review and cleanup
- [ ] Update documentation
- [ ] Team training on new architecture

## Risk Mitigation Strategies

### 1. **Feature Flags**
Use feature flags to toggle between old and new implementations:
```go
// internal/config/flags.go
type FeatureFlags struct {
    UseNewUserService   bool `env:"USE_NEW_USER_SERVICE" default:"false"`
    UseNewMemoryService bool `env:"USE_NEW_MEMORY_SERVICE" default:"false"`
}
```

### 2. **A/B Testing**
Run both implementations in parallel and compare results:
```go
// internal/migration/parallel.go
func (s *ParallelService) CreateUser(req user.CreateRequest) (*user.User, error) {
    // Run both implementations
    oldResult, oldErr := s.oldService.CreateUser(convertToOldReq(req))
    newResult, newErr := s.newService.CreateUser(context.Background(), req)
    
    // Compare results and log differences
    if !compareResults(oldResult, newResult) {
        log.Warn("Migration result mismatch", "req", req)
    }
    
    // Return old result during migration period
    return oldResult, oldErr
}
```

### 3. **Database Migration Safety**
- Use read replicas for testing
- Implement database schema versioning
- Plan for zero-downtime migrations

### 4. **Rollback Plan**
- Git tags for each migration phase
- Database rollback scripts
- Configuration rollback procedures
- Quick revert procedures documented

## Success Metrics

### 1. **Code Quality Metrics**
- Cyclomatic complexity reduction
- Test coverage improvement (target: >85%)
- Reduced coupling between packages
- Improved maintainability index

### 2. **Performance Metrics**
- Response time consistency
- Memory usage optimization
- CPU usage patterns
- Database query performance

### 3. **Development Metrics**
- Faster test execution
- Easier feature development
- Reduced bug introduction rate
- Improved developer experience

## Timeline Summary

| Phase | Duration | Risk Level | Key Deliverables |
|-------|----------|------------|------------------|
| 1. Preparation | 1-2 days | Low | New package structure, compatibility layer |
| 2. Domain/UseCase | 3-5 days | Medium | Clean domain layer, business logic |
| 3. Repository | 2-3 days | High | Data access layer, database safety |
| 4. Transport | 2-3 days | Medium | HTTP handlers, API compatibility |
| 5. Application | 1-2 days | Medium | Dependency injection, app wiring |
| 6. Cleanup | 2-3 days | Low | Remove old code, optimization |

**Total Estimated Time:** 2-3 weeks

The migration strategy is designed to minimize risk while providing maximum benefit. Each phase includes comprehensive testing and validation to ensure system stability throughout the transition.