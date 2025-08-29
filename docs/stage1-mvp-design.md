# Stage 1: MVP核心功能设计方案

## 阶段目标

Stage 1专注于建立一个可工作的记忆存储和检索MVP系统，基于现有的Go项目结构进行渐进式改进。这个阶段的目标是创建一个基础但完整的记忆银行，能够存储、检索和管理用户记忆。

### 核心功能范围

1. **基础记忆存储**：完善现有的记忆实体存储功能
2. **向量搜索**：集成pgvector实现语义搜索
3. **REST API**：提供完整的HTTP接口供外部调用
4. **嵌入服务**：集成BGE-M3模型生成向量嵌入
5. **基础测试**：确保功能正确性和稳定性

## 系统架构设计

### Stage 1架构图

```mermaid
graph TB
    subgraph "外部客户端 External Clients"
        CLIENT[AI应用客户端<br/>AI Application Client]
        POSTMAN[API测试工具<br/>Postman/curl]
    end
    
    subgraph "MCP API层 MCP API Layer"
        API[Memory API<br/>Gin HTTP Server]
        MIDDLEWARE[中间件层<br/>Auth, Logging, Validation]
    end
    
    subgraph "业务服务层 Business Service Layer"
        MEMORY_SVC[记忆服务<br/>Memory Service]
        EMBEDDING_SVC[嵌入服务<br/>Embedding Service]
        USER_SVC[用户服务<br/>User Service]
    end
    
    subgraph "数据访问层 Data Access Layer"
        MEMORY_REPO[记忆仓储<br/>Memory Repository]
        USER_REPO[用户仓储<br/>User Repository]
    end
    
    subgraph "存储层 Storage Layer"
        PG[(PostgreSQL<br/>+pgvector<br/>主要存储)]
        REDIS[(Redis<br/>缓存)]
    end
    
    subgraph "AI模型层 AI Model Layer"
        BGE[BGE-M3<br/>嵌入模型<br/>HTTP服务]
    end
    
    %% 请求流
    CLIENT --> API
    POSTMAN --> API
    API --> MIDDLEWARE
    MIDDLEWARE --> MEMORY_SVC
    MIDDLEWARE --> USER_SVC
    
    %% 服务调用
    MEMORY_SVC --> EMBEDDING_SVC
    MEMORY_SVC --> MEMORY_REPO
    USER_SVC --> USER_REPO
    
    %% 数据访问
    MEMORY_REPO --> PG
    MEMORY_REPO --> REDIS
    USER_REPO --> PG
    
    %% AI服务
    EMBEDDING_SVC --> BGE
    
    %% 样式
    style API fill:#87CEEB
    style MEMORY_SVC fill:#90EE90
    style EMBEDDING_SVC fill:#FFE4B5
    style PG fill:#DDA0DD
    style BGE fill:#F0E68C
```

## 数据模型设计

### 增强的记忆实体

基于现有的Memory实体，增加向量搜索所需的字段：

```go
// internal/domain/memory/entity.go (增强版本)
package memory

import (
    "time"
    "github.com/google/uuid"
    "mem_bank/internal/domain/user"
)

type Memory struct {
    ID           ID                     `json:"id" gorm:"type:uuid;primary_key"`
    UserID       user.ID                `json:"user_id" gorm:"type:uuid;not null;index"`
    Content      string                 `json:"content" gorm:"type:text;not null"`
    Summary      string                 `json:"summary" gorm:"type:text"`
    
    // 向量嵌入 (pgvector)
    Embedding    []float32              `json:"embedding" gorm:"type:vector(1024)"`
    
    // 记忆元数据
    Importance   int                    `json:"importance" gorm:"default:5"`
    MemoryType   MemoryType            `json:"memory_type" gorm:"type:varchar(50);not null"`
    Tags         []string              `json:"tags" gorm:"type:text[]"`
    Metadata     map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
    
    // 时间戳
    CreatedAt    time.Time             `json:"created_at"`
    UpdatedAt    time.Time             `json:"updated_at"`
    LastAccessed time.Time             `json:"last_accessed"`
    AccessCount  int                   `json:"access_count" gorm:"default:0"`
}

// 记忆类型枚举
type MemoryType string

const (
    MemoryTypeFactual   MemoryType = "factual"   // 事实性记忆
    MemoryTypeEpisodic  MemoryType = "episodic"  // 情景记忆
    MemoryTypePersonal  MemoryType = "personal"  // 个人偏好
    MemoryTypeProcedural MemoryType = "procedural" // 程序性记忆
)
```

### 数据库迁移

```sql
-- migrations/002_add_vector_support.up.sql
-- 启用 pgvector 扩展
CREATE EXTENSION IF NOT EXISTS vector;

-- 添加向量列到现有的 memories 表
ALTER TABLE memories ADD COLUMN embedding vector(1024);
ALTER TABLE memories ADD COLUMN memory_type varchar(50) DEFAULT 'factual';
ALTER TABLE memories ADD COLUMN tags text[];
ALTER TABLE memories ALTER COLUMN importance SET DEFAULT 5;

-- 创建向量搜索索引
CREATE INDEX ON memories USING hnsw (embedding vector_cosine_ops) 
WITH (m = 16, ef_construction = 64);

-- 为常用查询创建索引
CREATE INDEX idx_memories_user_id_type ON memories (user_id, memory_type);
CREATE INDEX idx_memories_created_at_desc ON memories (created_at DESC);
CREATE INDEX idx_memories_importance ON memories (importance DESC);

-- 为标签创建GIN索引支持数组查询
CREATE INDEX idx_memories_tags ON memories USING gin(tags);
```

## 核心服务实现

### 1. 嵌入服务增强

```go
// internal/service/embedding/service.go (增强版本)
package embedding

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "bytes"
    "time"
)

type Service struct {
    httpClient  *http.Client
    serviceURL  string
    modelName   string
}

type EmbeddingRequest struct {
    Texts     []string `json:"texts"`
    ModelName string   `json:"model_name"`
}

type EmbeddingResponse struct {
    Embeddings [][]float32 `json:"embeddings"`
    Model      string      `json:"model"`
    Usage      Usage       `json:"usage"`
}

type Usage struct {
    TotalTokens int `json:"total_tokens"`
}

func NewService(serviceURL, modelName string) *Service {
    return &Service{
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
        serviceURL: serviceURL,
        modelName:  modelName,
    }
}

func (s *Service) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
    if len(texts) == 0 {
        return nil, fmt.Errorf("no texts provided")
    }
    
    reqBody := EmbeddingRequest{
        Texts:     texts,
        ModelName: s.modelName,
    }
    
    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }
    
    req, err := http.NewRequestWithContext(ctx, "POST", 
        s.serviceURL+"/embeddings", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := s.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to call embedding service: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("embedding service returned status %d", resp.StatusCode)
    }
    
    var result EmbeddingResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    return result.Embeddings, nil
}

// 单文本嵌入便利方法
func (s *Service) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
    embeddings, err := s.GenerateEmbeddings(ctx, []string{text})
    if err != nil {
        return nil, err
    }
    
    if len(embeddings) == 0 {
        return nil, fmt.Errorf("no embedding returned")
    }
    
    return embeddings[0], nil
}
```

### 2. 记忆存储库实现

```go
// internal/dao/memory/vector_store.go (新文件)
package memory

import (
    "context"
    "fmt"
    "gorm.io/gorm"
    "mem_bank/internal/domain/memory"
    "mem_bank/internal/domain/user"
)

type PostgresVectorStore struct {
    db *gorm.DB
}

func NewPostgresVectorStore(db *gorm.DB) *PostgresVectorStore {
    return &PostgresVectorStore{db: db}
}

// 向量搜索方法
func (s *PostgresVectorStore) SearchSimilar(ctx context.Context, userID user.ID, 
    queryEmbedding []float32, limit int, threshold float64) ([]*memory.Memory, error) {
    
    var memories []*memory.Memory
    
    // 使用 pgvector 的余弦距离搜索
    // 注意: <-> 是余弦距离，<#> 是负内积，<=> 是欧氏距离
    query := `
        SELECT *, (embedding <-> ?) as distance 
        FROM memories 
        WHERE user_id = ? 
            AND embedding IS NOT NULL 
            AND (embedding <-> ?) < ?
        ORDER BY embedding <-> ? 
        LIMIT ?`
    
    // 将 []float32 转换为 pgvector 可识别的格式
    embeddingStr := vectorToString(queryEmbedding)
    
    err := s.db.WithContext(ctx).Raw(query, 
        embeddingStr, userID, embeddingStr, threshold, embeddingStr, limit).
        Scan(&memories).Error
    
    if err != nil {
        return nil, fmt.Errorf("failed to search similar memories: %w", err)
    }
    
    return memories, nil
}

// 批量向量搜索
func (s *PostgresVectorStore) BatchSearchSimilar(ctx context.Context, userID user.ID,
    queryEmbeddings [][]float32, limit int) ([][]*memory.Memory, error) {
    
    results := make([][]*memory.Memory, len(queryEmbeddings))
    
    for i, embedding := range queryEmbeddings {
        memories, err := s.SearchSimilar(ctx, userID, embedding, limit, 1.0)
        if err != nil {
            return nil, fmt.Errorf("batch search failed at index %d: %w", i, err)
        }
        results[i] = memories
    }
    
    return results, nil
}

// 按标签和相似度搜索
func (s *PostgresVectorStore) SearchByTagsAndSimilarity(ctx context.Context, 
    userID user.ID, tags []string, queryEmbedding []float32, limit int) ([]*memory.Memory, error) {
    
    var memories []*memory.Memory
    
    query := `
        SELECT *, (embedding <-> ?) as distance 
        FROM memories 
        WHERE user_id = ? 
            AND embedding IS NOT NULL
            AND tags && ?
        ORDER BY embedding <-> ? 
        LIMIT ?`
    
    embeddingStr := vectorToString(queryEmbedding)
    
    err := s.db.WithContext(ctx).Raw(query, 
        embeddingStr, userID, tags, embeddingStr, limit).
        Scan(&memories).Error
    
    if err != nil {
        return nil, fmt.Errorf("failed to search by tags and similarity: %w", err)
    }
    
    return memories, nil
}

// 工具函数：将 []float32 转换为 pgvector 字符串格式
func vectorToString(vector []float32) string {
    if len(vector) == 0 {
        return "[]"
    }
    
    result := "["
    for i, v := range vector {
        if i > 0 {
            result += ","
        }
        result += fmt.Sprintf("%.6f", v)
    }
    result += "]"
    return result
}
```

### 3. 记忆服务层实现

```go
// internal/service/memory/service.go (增强版本)
package memory

import (
    "context"
    "fmt"
    "mem_bank/internal/domain/memory"
    "mem_bank/internal/domain/user"
    "mem_bank/internal/service/embedding"
)

type Service struct {
    memoryRepo     memory.Repository
    embeddingRepo  memory.VectorStore
    embeddingService *embedding.Service
}

type CreateMemoryRequest struct {
    UserID     user.ID              `json:"user_id" validate:"required"`
    Content    string               `json:"content" validate:"required,min=1,max=5000"`
    Summary    string               `json:"summary,omitempty"`
    Importance int                  `json:"importance" validate:"min=1,max=10"`
    MemoryType memory.MemoryType    `json:"memory_type" validate:"required"`
    Tags       []string             `json:"tags,omitempty"`
    Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type SearchRequest struct {
    UserID      user.ID    `json:"user_id" validate:"required"`
    Query       string     `json:"query" validate:"required,min=1"`
    Limit       int        `json:"limit" validate:"min=1,max=50"`
    Tags        []string   `json:"tags,omitempty"`
    MemoryType  string     `json:"memory_type,omitempty"`
    Threshold   float64    `json:"threshold" validate:"min=0,max=2"`
}

func NewService(memoryRepo memory.Repository, vectorStore memory.VectorStore, 
    embeddingService *embedding.Service) *Service {
    return &Service{
        memoryRepo:       memoryRepo,
        embeddingRepo:    vectorStore,
        embeddingService: embeddingService,
    }
}

// 创建记忆（包含向量生成）
func (s *Service) CreateMemory(ctx context.Context, req *CreateMemoryRequest) (*memory.Memory, error) {
    // 生成向量嵌入
    embedding, err := s.embeddingService.GenerateEmbedding(ctx, req.Content)
    if err != nil {
        return nil, fmt.Errorf("failed to generate embedding: %w", err)
    }
    
    // 创建记忆实体
    mem := memory.NewMemory(req.UserID, req.Content, req.Summary, req.Importance, string(req.MemoryType))
    mem.Embedding = embedding
    mem.Tags = req.Tags
    mem.Metadata = req.Metadata
    
    // 存储到数据库
    if err := s.memoryRepo.Create(ctx, mem); err != nil {
        return nil, fmt.Errorf("failed to create memory: %w", err)
    }
    
    return mem, nil
}

// 语义搜索记忆
func (s *Service) SearchMemories(ctx context.Context, req *SearchRequest) ([]*memory.Memory, error) {
    // 生成查询向量
    queryEmbedding, err := s.embeddingService.GenerateEmbedding(ctx, req.Query)
    if err != nil {
        return nil, fmt.Errorf("failed to generate query embedding: %w", err)
    }
    
    // 设置默认阈值
    threshold := req.Threshold
    if threshold == 0 {
        threshold = 0.8 // 余弦相似度阈值
    }
    
    var memories []*memory.Memory
    
    // 根据是否有标签选择不同的搜索策略
    if len(req.Tags) > 0 {
        memories, err = s.embeddingRepo.SearchByTagsAndSimilarity(
            ctx, req.UserID, req.Tags, queryEmbedding, req.Limit)
    } else {
        memories, err = s.embeddingRepo.SearchSimilar(
            ctx, req.UserID, queryEmbedding, req.Limit, threshold)
    }
    
    if err != nil {
        return nil, fmt.Errorf("failed to search memories: %w", err)
    }
    
    // 记录访问
    for _, mem := range memories {
        mem.Access()
        // 异步更新访问统计
        go func(m *memory.Memory) {
            s.memoryRepo.Update(context.Background(), m)
        }(mem)
    }
    
    return memories, nil
}

// 获取用户的记忆统计
func (s *Service) GetMemoryStats(ctx context.Context, userID user.ID) (*MemoryStats, error) {
    stats := &MemoryStats{}
    
    // 总记忆数量
    total, err := s.memoryRepo.CountByUserID(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to count memories: %w", err)
    }
    stats.TotalMemories = total
    
    // 按类型分组统计
    typeStats, err := s.memoryRepo.CountByType(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to count by type: %w", err)
    }
    stats.ByType = typeStats
    
    return stats, nil
}

type MemoryStats struct {
    TotalMemories int                    `json:"total_memories"`
    ByType        map[string]int         `json:"by_type"`
    LastAccessed  *time.Time             `json:"last_accessed,omitempty"`
}
```

## API接口设计

### 记忆管理API

```go
// internal/handler/http/memory/handler.go (增强版本)
package memory

import (
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
    "mem_bank/internal/domain/user"
    "mem_bank/internal/service/memory"
)

type Handler struct {
    memoryService *memory.Service
}

func NewHandler(memoryService *memory.Service) *Handler {
    return &Handler{
        memoryService: memoryService,
    }
}

// POST /api/v1/memories
func (h *Handler) CreateMemory(c *gin.Context) {
    var req memory.CreateMemoryRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    mem, err := h.memoryService.CreateMemory(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, gin.H{
        "success": true,
        "data": mem,
    })
}

// POST /api/v1/memories/search
func (h *Handler) SearchMemories(c *gin.Context) {
    var req memory.SearchRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // 设置默认限制
    if req.Limit == 0 {
        req.Limit = 10
    }
    
    memories, err := h.memoryService.SearchMemories(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data": memories,
        "count": len(memories),
    })
}

// GET /api/v1/users/:user_id/memories/stats
func (h *Handler) GetMemoryStats(c *gin.Context) {
    userIDStr := c.Param("user_id")
    userID, err := user.ParseID(userIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
        return
    }
    
    stats, err := h.memoryService.GetMemoryStats(c.Request.Context(), userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data": stats,
    })
}

// GET /api/v1/users/:user_id/memories
func (h *Handler) GetUserMemories(c *gin.Context) {
    userIDStr := c.Param("user_id")
    userID, err := user.ParseID(userIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
        return
    }
    
    // 解析查询参数
    limitStr := c.DefaultQuery("limit", "20")
    limit, _ := strconv.Atoi(limitStr)
    if limit > 100 {
        limit = 100
    }
    
    offsetStr := c.DefaultQuery("offset", "0")
    offset, _ := strconv.Atoi(offsetStr)
    
    memoryType := c.Query("type")
    
    memories, err := h.memoryService.GetUserMemories(c.Request.Context(), userID, limit, offset, memoryType)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data": memories,
        "count": len(memories),
    })
}
```

## 部署配置

### BGE-M3嵌入服务

创建Python嵌入服务作为独立微服务：

```python
# scripts/embedding_service.py
from flask import Flask, request, jsonify
from FlagEmbedding import BGEM3FlagModel
import numpy as np
import logging

app = Flask(__name__)
logging.basicConfig(level=logging.INFO)

# 加载模型
model = None

def load_model():
    global model
    try:
        model = BGEM3FlagModel('./models/bge-m3', use_fp16=True)
        logging.info("BGE-M3 model loaded successfully")
    except Exception as e:
        logging.error(f"Failed to load model: {e}")
        raise

@app.route('/embeddings', methods=['POST'])
def generate_embeddings():
    try:
        data = request.json
        texts = data.get('texts', [])
        
        if not texts:
            return jsonify({"error": "No texts provided"}), 400
        
        # 生成嵌入
        embeddings = model.encode(texts, return_dense=True, return_sparse=False, return_colbert_vecs=False)
        dense_embeddings = embeddings['dense_vecs'].tolist()
        
        return jsonify({
            "embeddings": dense_embeddings,
            "model": "BAAI/bge-m3",
            "usage": {
                "total_tokens": sum(len(text.split()) for text in texts)
            }
        })
    
    except Exception as e:
        logging.error(f"Error generating embeddings: {e}")
        return jsonify({"error": str(e)}), 500

@app.route('/health', methods=['GET'])
def health_check():
    return jsonify({"status": "healthy", "model_loaded": model is not None})

if __name__ == '__main__':
    load_model()
    app.run(host='0.0.0.0', port=8001)
```

### Docker Compose配置

```yaml
# docker-compose.yml (更新)
version: '3.8'

services:
  # MCP主应用
  mcp_app:
    build: .
    container_name: mcp_app
    ports:
      - "8000:8000"
    volumes:
      - ./configs:/app/configs
    environment:
      - EMBEDDING_SERVICE_URL=http://embedding_service:8001
      - POSTGRES_DSN=postgres://user:password@postgres_db:5432/mcp_db?sslmode=disable
      - REDIS_ADDR=redis_cache:6379
    depends_on:
      - postgres_db
      - redis_cache
      - embedding_service
    restart: unless-stopped

  # BGE-M3嵌入服务
  embedding_service:
    build:
      context: .
      dockerfile: Dockerfile.embedding
    container_name: embedding_service
    ports:
      - "8001:8001"
    volumes:
      - ./models:/app/models
    deploy:
      resources:
        limits:
          memory: 4G
        reservations:
          memory: 2G
    restart: unless-stopped

  # PostgreSQL数据库
  postgres_db:
    image: pgvector/pgvector:pg16
    container_name: postgres_db
    environment:
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=mcp_db
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    ports:
      - "5432:5432"
    restart: unless-stopped

  # Redis缓存
  redis_cache:
    image: redis:7-alpine
    container_name: redis_cache
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
```

## 测试策略

### 单元测试示例

```go
// internal/service/memory/service_test.go
package memory_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    
    "mem_bank/internal/domain/memory"
    "mem_bank/internal/domain/user"
    "mem_bank/internal/service/memory"
)

func TestMemoryService_CreateMemory(t *testing.T) {
    // 设置模拟对象
    mockRepo := &MockMemoryRepository{}
    mockVectorStore := &MockVectorStore{}
    mockEmbedding := &MockEmbeddingService{}
    
    service := memory.NewService(mockRepo, mockVectorStore, mockEmbedding)
    
    // 模拟嵌入生成
    expectedEmbedding := make([]float32, 1024)
    for i := range expectedEmbedding {
        expectedEmbedding[i] = 0.1
    }
    mockEmbedding.On("GenerateEmbedding", mock.Anything, "test content").Return(expectedEmbedding, nil)
    
    // 模拟仓储创建
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*memory.Memory")).Return(nil)
    
    // 执行测试
    req := &memory.CreateMemoryRequest{
        UserID:     user.NewID(),
        Content:    "test content",
        Summary:    "test summary",
        Importance: 5,
        MemoryType: memory.MemoryTypeFactual,
    }
    
    mem, err := service.CreateMemory(context.Background(), req)
    
    // 断言
    assert.NoError(t, err)
    assert.NotNil(t, mem)
    assert.Equal(t, req.Content, mem.Content)
    assert.Equal(t, expectedEmbedding, mem.Embedding)
    
    mockEmbedding.AssertExpectations(t)
    mockRepo.AssertExpectations(t)
}
```

## 性能目标

### Stage 1性能指标

| 指标 | 目标值 | 测量方法 |
|------|--------|----------|
| **API响应时间** | P95 < 100ms | HTTP压测工具 |
| **向量搜索延迟** | < 50ms | 数据库查询日志 |
| **嵌入生成时间** | < 200ms (单文档) | 服务调用跟踪 |
| **并发处理能力** | 100 QPS | 负载测试 |
| **内存使用** | < 512MB (不含模型) | 系统监控 |
| **数据库连接** | < 10个活跃连接 | PostgreSQL监控 |

## 里程碑和验收标准

### M1.1: 基础存储功能 (3天)

**完成标准：**

- ✅ 数据库迁移脚本完成
- ✅ 记忆实体CRUD操作正常
- ✅ pgvector索引创建成功
- ✅ 基础单元测试通过

### M1.2: 向量搜索集成 (2天)

**完成标准：**

- ✅ BGE-M3服务正常运行
- ✅ 嵌入生成API可用
- ✅ 向量相似度搜索功能正常
- ✅ 搜索结果准确性验证

### M1.3: REST API完善 (2天)

**完成标准：**

- ✅ 所有API端点实现
- ✅ 请求验证和错误处理
- ✅ API文档完整
- ✅ Postman测试集合

### M1.4: 集成测试和部署 (1天)

**完成标准：**

- ✅ Docker Compose部署成功
- ✅ 端到端测试通过
- ✅ 性能指标达标
- ✅ 文档和示例完整

## 下一阶段预览

Stage 1完成后，将为Stage 2的智能记忆处理引擎奠定坚实基础：

- LLM集成和事实提取
- 记忆整合算法（ADD/UPDATE/DELETE/NOOP）
- 语境合成和提示优化
- 异步处理和队列机制

Stage 1的MVP将证明基础架构的可行性，为后续的智能化功能提供稳定的技术底座。