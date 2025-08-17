# RFC: AI记忆系统第一阶段MVP设计

**文档版本**: 1.0  
**创建日期**: 2025-08-17  
**作者**: AI记忆系统团队  
**状态**: 草案

## 1. 摘要

本文档描述AI记忆系统第一阶段最小可行产品(MVP)的详细设计。该阶段专注于建立核心基础架构，包括领域模型、配置管理、数据存储层和基础API功能。采用Clean Architecture原则，为后续阶段的智能化功能奠定坚实基础。

## 2. 引言

### 2.1 目标

- 建立可扩展的项目基础架构
- 实现基础的记忆CRUD操作
- 集成PostgreSQL+pgvector作为主要存储后端
- 提供RESTful API接口
- 实现基础的LLM集成能力

### 2.2 范围

第一阶段包含以下功能：
- 核心领域模型定义
- 配置管理系统
- PostgreSQL数据库集成
- 基础HTTP API
- 简单的向量嵌入生成

### 2.3 术语

- **MVP**: 最小可行产品
- **Clean Architecture**: 整洁架构模式
- **pgvector**: PostgreSQL向量扩展
- **Repository Pattern**: 仓库模式
- **Domain-Driven Design**: 领域驱动设计

## 3. 系统架构

### 3.1 整体架构

```mermaid
graph TB
    subgraph "外部客户端"
        A[AI Agent]
        B[Web Client]
    end
    
    subgraph "API层"
        C[Gin Router]
        D[HTTP Handlers]
    end
    
    subgraph "业务逻辑层"
        E[Memory Usecase]
        F[Domain Services]
    end
    
    subgraph "数据访问层"
        G[Memory Repository Interface]
        H[PostgreSQL Repository Impl]
    end
    
    subgraph "基础设施层"
        I[PostgreSQL Database]
        J[pgvector Extension]
        K[Configuration]
        L[Logger]
    end
    
    subgraph "外部服务"
        M[LLM Provider]
    end
    
    A --> C
    B --> C
    C --> D
    D --> E
    E --> G
    G --> H
    H --> I
    I --> J
    E --> M
    
    style E fill:#e1f5fe
    style G fill:#f3e5f5
    style K fill:#fff3e0
```

### 3.2 分层架构详解

#### 3.2.1 API层 (Delivery Layer)
- **职责**: 处理HTTP请求，路由分发，数据验证
- **技术**: Gin框架
- **输入**: HTTP请求
- **输出**: JSON响应

#### 3.2.2 业务逻辑层 (Use Case Layer)  
- **职责**: 核心业务逻辑实现，用例编排
- **技术**: Go接口和结构体
- **依赖**: Repository接口，Domain模型

#### 3.2.3 数据访问层 (Repository Layer)
- **职责**: 数据持久化抽象
- **技术**: PostgreSQL + pgvector
- **模式**: Repository Pattern

#### 3.2.4 基础设施层 (Infrastructure Layer)
- **职责**: 外部依赖，配置，日志
- **技术**: Viper, logrus, pgx

## 4. 详细设计

### 4.1 领域模型

#### 4.1.1 核心实体

```go
// Memory 记忆实体
type Memory struct {
    ID        string                 `json:"id" db:"id"`
    UserID    string                 `json:"user_id" db:"user_id"`
    Content   string                 `json:"content" db:"content"`
    Embedding []float32              `json:"-" db:"embedding"`
    Metadata  map[string]interface{} `json:"metadata" db:"metadata"`
    CreatedAt time.Time              `json:"created_at" db:"created_at"`
    UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
}

// User 用户实体
type User struct {
    ID        string    `json:"id" db:"id"`
    Name      string    `json:"name" db:"name"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}
```

#### 4.1.2 业务接口

```go
// MemoryUsecase 记忆用例接口
type MemoryUsecase interface {
    Create(ctx context.Context, userID string, content string) (*Memory, error)
    GetByID(ctx context.Context, memoryID string) (*Memory, error)
    GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*Memory, error)
    Update(ctx context.Context, memoryID string, content string) (*Memory, error)
    Delete(ctx context.Context, memoryID string) error
    Search(ctx context.Context, userID string, query string, limit int) ([]*Memory, error)
}

// MemoryRepository 记忆仓库接口
type MemoryRepository interface {
    Save(ctx context.Context, memory *Memory) error
    FindByID(ctx context.Context, id string) (*Memory, error)
    FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*Memory, error)
    Update(ctx context.Context, memory *Memory) error
    Delete(ctx context.Context, id string) error
    FindSimilar(ctx context.Context, userID string, embedding []float32, limit int) ([]*Memory, error)
}

// LLMProvider LLM服务接口
type LLMProvider interface {
    CreateEmbedding(ctx context.Context, text string) ([]float32, error)
    CreateEmbeddings(ctx context.Context, texts []string) ([][]float32, error)
}
```

### 4.2 数据库设计

#### 4.2.1 数据库模式

```sql
-- 启用pgvector扩展
CREATE EXTENSION IF NOT EXISTS vector;

-- 用户表
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 记忆表
CREATE TABLE memories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    embedding VECTOR(1536) NOT NULL, -- OpenAI ada-002 embedding维度
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_memories_user_id ON memories(user_id);
CREATE INDEX idx_memories_created_at ON memories(created_at DESC);

-- 向量索引 (HNSW)
CREATE INDEX idx_memories_embedding_hnsw ON memories USING hnsw (embedding vector_cosine_ops);
```

#### 4.2.2 数据库迁移流程

```mermaid
flowchart TD
    A[启动应用] --> B[检查迁移状态]
    B --> C{是否需要迁移?}
    C -->|是| D[执行迁移脚本]
    C -->|否| E[启动应用服务]
    D --> F{迁移成功?}
    F -->|是| E
    F -->|否| G[记录错误并退出]
    
    style D fill:#fff3e0
    style F fill:#e8f5e8
    style G fill:#ffebee
```

### 4.3 API设计

#### 4.3.1 REST API端点

| 方法 | 路径 | 描述 | 请求体 | 响应 |
|------|------|------|--------|------|
| POST | `/api/v1/memories` | 创建记忆 | `CreateMemoryRequest` | `Memory` |
| GET | `/api/v1/memories/{id}` | 获取记忆 | - | `Memory` |
| GET | `/api/v1/users/{user_id}/memories` | 获取用户记忆列表 | - | `MemoryList` |
| PUT | `/api/v1/memories/{id}` | 更新记忆 | `UpdateMemoryRequest` | `Memory` |
| DELETE | `/api/v1/memories/{id}` | 删除记忆 | - | - |
| POST | `/api/v1/memories/search` | 搜索记忆 | `SearchRequest` | `MemoryList` |

#### 4.3.2 数据传输对象

```go
// CreateMemoryRequest 创建记忆请求
type CreateMemoryRequest struct {
    UserID   string                 `json:"user_id" binding:"required,uuid"`
    Content  string                 `json:"content" binding:"required,min=1,max=10000"`
    Metadata map[string]interface{} `json:"metadata"`
}

// UpdateMemoryRequest 更新记忆请求
type UpdateMemoryRequest struct {
    Content  string                 `json:"content" binding:"required,min=1,max=10000"`
    Metadata map[string]interface{} `json:"metadata"`
}

// SearchRequest 搜索请求
type SearchRequest struct {
    UserID string `json:"user_id" binding:"required,uuid"`
    Query  string `json:"query" binding:"required,min=1,max=1000"`
    Limit  int    `json:"limit" binding:"min=1,max=100"`
}

// MemoryList 记忆列表响应
type MemoryList struct {
    Memories []*Memory `json:"memories"`
    Total    int       `json:"total"`
    Limit    int       `json:"limit"`
    Offset   int       `json:"offset"`
}
```

### 4.4 配置管理

#### 4.4.1 配置结构

```go
type Config struct {
    // 服务配置
    Server ServerConfig `mapstructure:"server"`
    
    // 数据库配置
    Database DatabaseConfig `mapstructure:"database"`
    
    // LLM配置
    LLM LLMConfig `mapstructure:"llm"`
    
    // 日志配置
    Log LogConfig `mapstructure:"log"`
}

type ServerConfig struct {
    Port         string        `mapstructure:"port"`
    ReadTimeout  time.Duration `mapstructure:"read_timeout"`
    WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
    URL         string `mapstructure:"url"`
    MaxOpenConn int    `mapstructure:"max_open_conn"`
    MaxIdleConn int    `mapstructure:"max_idle_conn"`
}

type LLMConfig struct {
    Provider string `mapstructure:"provider"` // "openai"
    APIKey   string `mapstructure:"api_key"`
    BaseURL  string `mapstructure:"base_url"`
}

type LogConfig struct {
    Level  string `mapstructure:"level"`
    Format string `mapstructure:"format"` // "json" or "text"
}
```

#### 4.4.2 配置加载流程

```mermaid
sequenceDiagram
    participant App as 应用启动
    participant Viper as Viper
    participant File as 配置文件
    participant Env as 环境变量
    participant Config as 配置结构

    App->>Viper: 初始化配置管理
    Viper->>File: 读取config.yaml
    File-->>Viper: 返回文件配置
    Viper->>Env: 读取环境变量
    Env-->>Viper: 返回环境配置
    Viper->>Viper: 合并配置(环境变量优先)
    Viper->>Config: 反序列化到结构体
    Config-->>App: 返回配置对象
```

### 4.5 实现流程

#### 4.5.1 创建记忆流程

```mermaid
sequenceDiagram
    participant Client as 客户端
    participant Handler as HTTP Handler
    participant Usecase as Memory Usecase
    participant LLM as LLM Provider
    participant Repo as Repository
    participant DB as PostgreSQL

    Client->>Handler: POST /api/v1/memories
    Handler->>Handler: 验证请求数据
    Handler->>Usecase: Create(userID, content)
    
    Usecase->>LLM: CreateEmbedding(content)
    LLM-->>Usecase: 返回向量嵌入
    
    Usecase->>Usecase: 构建Memory对象
    Usecase->>Repo: Save(memory)
    
    Repo->>DB: INSERT INTO memories
    DB-->>Repo: 返回记录ID
    Repo-->>Usecase: 返回完整记忆对象
    Usecase-->>Handler: 返回记忆对象
    Handler-->>Client: 201 Created + Memory JSON
```

#### 4.5.2 搜索记忆流程

```mermaid
sequenceDiagram
    participant Client as 客户端
    participant Handler as HTTP Handler
    participant Usecase as Memory Usecase
    participant LLM as LLM Provider
    participant Repo as Repository
    participant DB as PostgreSQL

    Client->>Handler: POST /api/v1/memories/search
    Handler->>Handler: 验证搜索参数
    Handler->>Usecase: Search(userID, query, limit)
    
    Usecase->>LLM: CreateEmbedding(query)
    LLM-->>Usecase: 返回查询向量
    
    Usecase->>Repo: FindSimilar(userID, embedding, limit)
    
    Repo->>DB: SELECT ... ORDER BY embedding <=> $1 LIMIT $2
    DB-->>Repo: 返回相似记忆列表
    Repo-->>Usecase: 返回记忆列表
    Usecase-->>Handler: 返回搜索结果
    Handler-->>Client: 200 OK + MemoryList JSON
```

## 5. 实施计划

### 5.1 开发里程碑

#### 第1周: 项目基础设施
- [ ] 项目结构搭建
- [ ] Go模块初始化
- [ ] 依赖管理配置
- [ ] 开发环境设置

#### 第2周: 核心模型与配置
- [ ] 领域模型定义
- [ ] 配置管理实现
- [ ] 日志系统集成
- [ ] 数据库连接

#### 第3周: 数据访问层
- [ ] PostgreSQL集成
- [ ] pgvector配置
- [ ] Repository实现
- [ ] 数据库迁移

#### 第4周: 业务逻辑层
- [ ] Usecase实现
- [ ] LLM集成
- [ ] 向量嵌入生成
- [ ] 单元测试

#### 第5周: API层
- [ ] Gin框架集成
- [ ] HTTP处理器
- [ ] 路由配置
- [ ] 请求验证

#### 第6周: 集成测试与优化
- [ ] 集成测试
- [ ] API测试
- [ ] 性能调优
- [ ] 文档完善

### 5.2 技术债务管理

#### 已知限制
1. **同步处理**: 当前阶段使用同步处理，后续需要改为异步
2. **单一数据库**: 仅支持PostgreSQL，需要抽象化支持多数据库
3. **简单LLM集成**: 仅支持基础嵌入生成，缺少工具调用能力
4. **基础错误处理**: 错误处理较为简单，需要增强

#### 重构计划
- 第二阶段：引入异步处理机制
- 第三阶段：实现多数据库支持
- 第四阶段：完善监控和可观测性

## 6. 测试策略

### 6.1 单元测试

#### 测试覆盖范围
- 领域模型验证
- 业务逻辑测试
- Repository层测试
- API处理器测试

#### 测试工具
- `testify`: 断言和模拟框架
- `go-mock`: 接口模拟生成
- `dockertest`: 数据库集成测试

### 6.2 集成测试

#### 测试场景
- 端到端API流程
- 数据库操作验证
- LLM服务集成
- 配置加载测试

### 6.3 测试数据管理

```mermaid
graph LR
    A[测试启动] --> B[创建测试数据库]
    B --> C[运行迁移]
    C --> D[插入测试数据]
    D --> E[执行测试]
    E --> F[清理测试数据]
    F --> G[删除测试数据库]
    
    style B fill:#e8f5e8
    style F fill:#fff3e0
```

## 7. 部署与运维

### 7.1 容器化

#### Dockerfile
```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o mem-bank ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/mem-bank .
COPY --from=builder /app/configs ./configs
EXPOSE 8080
CMD ["./mem-bank"]
```

#### Docker Compose
```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://user:pass@postgres:5432/membank?sslmode=disable
    depends_on:
      - postgres

  postgres:
    image: pgvector/pgvector:pg16
    environment:
      - POSTGRES_DB=membank
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=pass
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

volumes:
  postgres_data:
```

### 7.2 健康检查

#### 健康检查端点
```go
func (h *HealthHandler) Check(c *gin.Context) {
    status := map[string]string{
        "status": "healthy",
        "database": h.checkDatabase(),
        "llm": h.checkLLM(),
    }
    c.JSON(http.StatusOK, status)
}
```

## 8. 安全考虑

### 8.1 数据安全
- 数据库连接加密
- 敏感配置环境变量化
- API密钥安全存储

### 8.2 输入验证
- 请求参数验证
- SQL注入防护
- XSS防护

### 8.3 访问控制
- 用户身份验证（预留接口）
- 记忆访问权限控制
- API限流保护

## 9. 总结

第一阶段MVP设计提供了AI记忆系统的基础架构和核心功能。通过采用Clean Architecture和Repository模式，为系统的可扩展性和可维护性奠定了坚实基础。后续阶段将在此基础上构建智能化的记忆管理能力。

### 9.1 交付成果
- 完整的CRUD API
- PostgreSQL+pgvector集成
- 基础向量搜索能力
- 配置管理系统
- 容器化部署方案

### 9.2 后续阶段准备
- 预留异步处理接口
- 抽象数据库访问层
- 设计扩展点为智能化功能做准备

---
**文档状态**: 草案  
**下次评审**: 2025-08-24  
**相关文档**: [总体架构设计](./overall-architecture.md), [第二阶段设计](./stage2-intelligent-design.md)