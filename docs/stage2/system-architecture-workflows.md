# AI 记忆银行系统架构与工作流程图

**文档版本**: 1.0  
**创建日期**: 2025-08-26  
**状态**: 系统架构文档  

## 1. 系统整体架构图

### 1.1 分层架构概览

```mermaid
graph TB
    subgraph "客户端层 (Client Layer)"
        A[AI Agent客户端]
        B[Web应用]
        C[移动应用]
        D[第三方集成]
    end
    
    subgraph "API网关层 (API Gateway Layer)"
        E[Gin HTTP Router]
        F[认证中间件]
        G[日志中间件]
        H[错误处理中间件]
        I[CORS中间件]
    end
    
    subgraph "应用服务层 (Application Service Layer)"
        J[Memory Handler]
        K[User Handler]
        L[Admin Handler]
    end
    
    subgraph "业务逻辑层 (Business Logic Layer)"
        M[AI Memory Service]
        N[User Service]
        O[Embedding Service]
        P[Queue Manager]
    end
    
    subgraph "基础设施层 (Infrastructure Layer)"
        Q[LLM Provider]
        R[Job Queue System]
        S[Cache Manager]
        T[Configuration]
    end
    
    subgraph "数据持久化层 (Data Persistence Layer)"
        U[Memory Repository]
        V[User Repository]
        W[(PostgreSQL + pgvector)]
        X[(Qdrant Vector DB)]
        Y[(Redis Cache & Queue)]
    end
    
    A --> E
    B --> E
    C --> E
    D --> E
    
    E --> F
    F --> G
    G --> H
    H --> I
    I --> J
    I --> K
    I --> L
    
    J --> M
    K --> N
    L --> N
    
    M --> O
    M --> P
    M --> U
    N --> V
    
    O --> Q
    O --> S
    P --> R
    R --> Y
    
    U --> W
    U --> X
    V --> W
    S --> Y
    
    style M fill:#e3f2fd
    style O fill:#f3e5f5
    style Q fill:#fff3e0
    style R fill:#e8f5e8
```

### 1.2 核心组件依赖关系

```mermaid
graph LR
    subgraph "Domain Layer"
        A[Memory Entity]
        B[User Entity]
        C[Service Interfaces]
    end
    
    subgraph "Application Layer"
        D[AI Memory Service]
        E[Embedding Service]
        F[User Service]
    end
    
    subgraph "Infrastructure Layer"
        G[PostgreSQL Repository]
        H[Qdrant Repository]
        I[Redis Queue]
        J[OpenAI Provider]
    end
    
    D --> A
    D --> C
    E --> J
    F --> B
    F --> C
    
    D --> G
    D --> H
    D --> I
    E --> I
    
    C -.-> G
    C -.-> H
    
    style A fill:#e3f2fd
    style D fill:#f3e5f5
    style E fill:#fff3e0
```

## 2. 核心工作流程图

### 2.1 记忆创建完整流程

```mermaid
sequenceDiagram
    participant Client as 客户端
    participant API as Gin Router
    participant Handler as Memory Handler
    participant Service as AI Memory Service
    participant Embed as Embedding Service
    participant Queue as Job Queue
    participant LLM as LLM Provider
    participant Cache as Redis Cache
    participant DB as Vector Database

    Client->>+API: POST /api/v1/memories
    API->>+Handler: CreateMemory(request)
    Handler->>Handler: 验证请求参数
    Handler->>+Service: CreateMemory(req)
    
    Service->>Service: 验证用户存在
    Service->>Service: 创建Memory实体
    
    alt 同步嵌入生成模式
        Service->>+Embed: GenerateEmbedding(content)
        Embed->>+Cache: 检查嵌入缓存
        Cache-->>-Embed: 缓存未命中
        Embed->>+LLM: GenerateEmbeddings(content)
        LLM-->>-Embed: 返回嵌入向量
        Embed->>Cache: 缓存嵌入结果
        Embed-->>-Service: 返回嵌入向量
        Service->>Service: 更新Memory.Embedding
    else 异步嵌入生成模式
        Service->>+Queue: EnqueueEmbeddingJob(memoryID)
        Queue-->>-Service: 返回JobID
        Note over Queue: 后台异步处理
    end
    
    Service->>+DB: Store(memory)
    DB-->>-Service: 存储成功
    Service-->>-Handler: 返回创建的Memory
    Handler-->>-API: JSON响应
    API-->>-Client: 201 Created + Memory数据
    
    Note over Queue,DB: 异步嵌入生成流程
    Queue->>+Embed: ProcessEmbeddingJob
    Embed->>+LLM: GenerateEmbeddings
    LLM-->>-Embed: 嵌入向量
    Embed->>+DB: UpdateMemoryEmbedding
    DB-->>-Embed: 更新成功
    Embed-->>-Queue: 作业完成
```

### 2.2 语义相似性搜索流程

```mermaid
sequenceDiagram
    participant Client as 客户端
    participant API as API Gateway
    participant Handler as Memory Handler
    participant Service as AI Memory Service
    participant Embed as Embedding Service
    participant Cache as Redis Cache
    participant LLM as LLM Provider
    participant VectorDB as Vector Database

    Client->>+API: GET /api/v1/memories/users/{id}/similar?content=...
    API->>+Handler: SearchSimilarMemories
    Handler->>Handler: 参数验证和解析
    Handler->>+Service: SearchSimilarMemories(content, userID, limit, threshold)
    
    Service->>+Embed: GenerateEmbedding(searchContent)
    Embed->>+Cache: 检查查询内容嵌入缓存
    
    alt 缓存命中
        Cache-->>Embed: 返回缓存的嵌入向量
    else 缓存未命中  
        Embed->>+LLM: GenerateEmbeddings(content)
        LLM-->>-Embed: 嵌入向量
        Embed->>Cache: 缓存嵌入结果
    end
    
    Embed-->>-Service: 查询嵌入向量
    
    Service->>+VectorDB: SearchSimilar(embedding, userID, limit, threshold)
    VectorDB->>VectorDB: 执行向量相似性搜索
    VectorDB->>VectorDB: 应用相似度阈值过滤
    VectorDB->>VectorDB: 用户数据权限过滤
    VectorDB-->>-Service: 相似记忆列表
    
    Service->>Service: 更新访问统计
    Service-->>-Handler: 搜索结果
    Handler-->>-API: JSON响应
    API-->>-Client: 200 OK + 相似记忆列表
```

### 2.3 混合搜索算法流程

```mermaid
flowchart TD
    A[搜索请求] --> B{查询类型判断}
    B -->|纯文本查询| C[文本搜索]
    B -->|语义查询| D[语义搜索]
    B -->|混合查询| E[并行搜索]
    
    E --> F[文本关键词搜索]
    E --> G[语义向量搜索]
    
    F --> H[文本搜索结果]
    G --> I[语义搜索结果]
    
    H --> J[结果合并算法]
    I --> J
    
    J --> K[去重处理]
    K --> L[权重计算]
    L --> M[综合排序]
    M --> N[分页和限制]
    N --> O[返回最终结果]
    
    C --> P[直接文本匹配]
    D --> Q[纯语义匹配]
    P --> O
    Q --> O
    
    style E fill:#e3f2fd
    style J fill:#f3e5f5
    style M fill:#fff3e0
```

### 2.4 异步作业处理流程

```mermaid
graph TD
    A[作业产生] --> B[Job Queue]
    B --> C{作业类型}
    
    C -->|generate_embedding| D[嵌入生成作业]
    C -->|batch_embedding| E[批量嵌入作业]
    C -->|其他作业类型| F[扩展作业处理]
    
    D --> G[Worker Pool]
    E --> G
    F --> G
    
    G --> H[作业执行]
    H --> I{执行结果}
    
    I -->|成功| J[标记完成]
    I -->|失败| K{重试次数检查}
    
    K -->|未达上限| L[延迟重试]
    K -->|达到上限| M[标记失败]
    
    L --> B
    J --> N[作业统计]
    M --> N
    
    N --> O[清理过期作业]
    
    style B fill:#e3f2fd
    style G fill:#f3e5f5
    style H fill:#fff3e0
    style N fill:#e8f5e8
```

## 3. 数据流架构图

### 3.1 记忆数据生命周期

```mermaid
stateDiagram-v2
    [*] --> Created: 用户创建记忆
    Created --> EmbeddingQueued: 排队生成嵌入
    Created --> EmbeddingGenerated: 同步生成嵌入
    
    EmbeddingQueued --> EmbeddingGenerating: Worker开始处理
    EmbeddingGenerating --> EmbeddingGenerated: 嵌入生成成功
    EmbeddingGenerating --> EmbeddingFailed: 嵌入生成失败
    EmbeddingFailed --> EmbeddingQueued: 重试
    
    EmbeddingGenerated --> Searchable: 可被搜索
    Searchable --> Accessed: 用户访问
    Accessed --> Searchable: 更新访问统计
    
    Searchable --> Updated: 用户更新内容
    Updated --> EmbeddingQueued: 重新生成嵌入
    
    Searchable --> Deleted: 用户删除
    Deleted --> [*]
    
    EmbeddingFailed --> Deleted: 多次失败后删除
```

### 3.2 嵌入缓存策略

```mermaid
graph LR
    A[文本内容] --> B[内容哈希]
    B --> C{Redis缓存检查}
    
    C -->|命中| D[返回缓存嵌入]
    C -->|未命中| E[调用LLM API]
    
    E --> F[生成嵌入向量]
    F --> G[存储到Redis]
    G --> H[设置TTL过期时间]
    H --> I[返回嵌入向量]
    
    D --> J[更新缓存访问统计]
    I --> K[记录API调用统计]
    
    J --> L[嵌入结果]
    K --> L
    
    style C fill:#e3f2fd
    style G fill:#f3e5f5
    style L fill:#e8f5e8
```

## 4. 系统交互序列图

### 4.1 用户认证和权限检查

```mermaid
sequenceDiagram
    participant Client as 客户端
    participant Gateway as API Gateway
    participant Auth as 认证中间件
    participant Handler as 业务处理器
    participant Service as 业务服务

    Client->>+Gateway: 请求 + API Key
    Gateway->>+Auth: 验证请求
    
    Auth->>Auth: 提取API Key
    Auth->>Auth: 验证API Key有效性
    
    alt API Key有效
        Auth->>Auth: 设置用户上下文
        Auth->>+Handler: 转发请求
        Handler->>+Service: 调用业务逻辑
        Service->>Service: 检查用户权限
        Service-->>-Handler: 业务结果
        Handler-->>-Auth: 处理结果
        Auth-->>-Gateway: 成功响应
    else API Key无效
        Auth-->>-Gateway: 401 Unauthorized
    end
    
    Gateway-->>-Client: 最终响应
```

### 4.2 错误处理和恢复流程

```mermaid
sequenceDiagram
    participant Client as 客户端
    participant API as API层
    participant Service as 服务层
    participant External as 外部服务
    participant Logger as 日志系统
    participant Monitor as 监控系统

    Client->>+API: 业务请求
    API->>+Service: 调用服务
    Service->>+External: 调用外部API
    
    alt 外部服务成功
        External-->>Service: 正常响应
        Service-->>API: 业务结果
        API-->>Client: 成功响应
    else 外部服务失败
        External-->>-Service: 错误响应
        Service->>Service: 错误分类和处理
        Service->>+Logger: 记录错误日志
        Logger-->>-Service: 日志记录完成
        Service->>+Monitor: 发送错误指标
        Monitor-->>-Service: 指标记录完成
        
        alt 可重试错误
            Service->>Service: 重试逻辑
            Service->>External: 重试请求
            External-->>Service: 重试结果
        else 不可重试错误
            Service->>Service: 错误包装
        end
        
        Service-->>-API: 错误响应
        API->>API: 错误格式标准化
        API-->>-Client: 标准化错误响应
    end
```

## 5. 配置和部署架构

### 5.1 配置管理流程

```mermaid
flowchart TD
    A[应用启动] --> B[配置加载器]
    B --> C[环境变量]
    B --> D[配置文件]
    B --> E[命令行参数]
    B --> F[默认值]
    
    C --> G[Viper配置合并]
    D --> G
    E --> G
    F --> G
    
    G --> H[配置验证]
    H --> I{验证通过?}
    
    I -->|是| J[配置对象创建]
    I -->|否| K[启动失败]
    
    J --> L[依赖注入配置]
    L --> M[服务初始化]
    M --> N[应用运行]
    
    K --> O[错误日志]
    O --> P[退出程序]
    
    style G fill:#e3f2fd
    style H fill:#f3e5f5
    style L fill:#fff3e0
```

### 5.2 容器化部署架构

```mermaid
graph TB
    subgraph "开发环境"
        A[源代码]
        B[单元测试]
        C[集成测试]
    end
    
    subgraph "CI/CD管道"
        D[代码提交]
        E[自动化测试]
        F[Docker构建]
        G[镜像推送]
    end
    
    subgraph "生产环境"
        H[负载均衡器]
        I[应用实例1]
        J[应用实例2]
        K[应用实例N]
    end
    
    subgraph "数据层"
        L[(PostgreSQL主)]
        M[(PostgreSQL从)]
        N[(Redis集群)]
        O[(Qdrant集群)]
    end
    
    subgraph "监控系统"
        P[Prometheus]
        Q[Grafana]
        R[AlertManager]
    end
    
    A --> D
    B --> E
    C --> E
    D --> E
    E --> F
    F --> G
    G --> H
    
    H --> I
    H --> J
    H --> K
    
    I --> L
    I --> M
    I --> N
    I --> O
    
    J --> L
    J --> M
    J --> N
    J --> O
    
    K --> L
    K --> M
    K --> N
    K --> O
    
    I --> P
    J --> P
    K --> P
    P --> Q
    P --> R
    
    style F fill:#e3f2fd
    style H fill:#f3e5f5
    style P fill:#fff3e0
```

## 6. 性能监控架构

### 6.1 监控指标收集流程

```mermaid
graph TD
    A[应用实例] --> B[指标收集器]
    B --> C[业务指标]
    B --> D[系统指标]
    B --> E[HTTP指标]
    B --> F[数据库指标]
    
    C --> G[Prometheus]
    D --> G
    E --> G
    F --> G
    
    G --> H[指标存储]
    H --> I[Grafana仪表板]
    H --> J[告警规则引擎]
    
    J --> K{阈值检查}
    K -->|正常| L[继续监控]
    K -->|异常| M[触发告警]
    
    M --> N[通知渠道]
    N --> O[运维团队]
    
    L --> G
    
    style B fill:#e3f2fd
    style G fill:#f3e5f5
    style J fill:#fff3e0
    style M fill:#ff6b6b
```

### 6.2 日志聚合架构

```mermaid
flowchart LR
    subgraph "应用层"
        A[App Instance 1]
        B[App Instance 2] 
        C[App Instance N]
    end
    
    subgraph "日志收集"
        D[Filebeat/Fluentd]
        E[日志缓冲]
    end
    
    subgraph "日志处理"
        F[Logstash/Fluentd]
        G[日志解析]
        H[日志过滤]
        I[日志格式化]
    end
    
    subgraph "日志存储"
        J[Elasticsearch]
        K[日志索引]
    end
    
    subgraph "日志可视化"
        L[Kibana]
        M[日志搜索]
        N[仪表板]
    end
    
    A --> D
    B --> D
    C --> D
    D --> E
    E --> F
    F --> G
    G --> H
    H --> I
    I --> J
    J --> K
    K --> L
    L --> M
    L --> N
    
    style F fill:#e3f2fd
    style J fill:#f3e5f5
    style L fill:#fff3e0
```

## 7. 数据库架构设计

### 7.1 PostgreSQL + pgvector架构

```mermaid
erDiagram
    USERS ||--o{ MEMORIES : owns
    USERS {
        uuid id PK
        string username UK
        string email UK
        jsonb profile
        boolean is_active
        timestamp created_at
        timestamp updated_at
        timestamp last_login
    }
    
    MEMORIES ||--o{ MEMORY_TAGS : has
    MEMORIES {
        uuid id PK
        uuid user_id FK
        text content
        text summary
        vector embedding "pgvector(1536)"
        integer importance
        string memory_type
        jsonb metadata
        timestamp created_at
        timestamp updated_at
        timestamp last_accessed
        integer access_count
    }
    
    TAGS ||--o{ MEMORY_TAGS : referenced_by
    TAGS {
        uuid id PK
        string name UK
        string description
        timestamp created_at
    }
    
    MEMORY_TAGS {
        uuid memory_id FK
        uuid tag_id FK
    }
    
    MEMORY_JOBS {
        uuid id PK
        string job_type
        uuid memory_id FK
        string status
        integer priority
        integer retry_count
        jsonb payload
        text error_message
        timestamp created_at
        timestamp started_at
        timestamp completed_at
    }
```

### 7.2 Qdrant向量数据库架构

```mermaid
graph TD
    A[Qdrant Collection: memories] --> B[Point Structure]
    B --> C[ID: Memory UUID]
    B --> D[Vector: Embedding 1536D]
    B --> E[Payload: Metadata]
    
    E --> F[user_id: UUID]
    E --> G[content: Text]
    E --> H[memory_type: String]
    E --> I[importance: Integer]
    E --> J[tags: Array]
    E --> K[created_at: Timestamp]
    
    A --> L[Index Configuration]
    L --> M[Distance: Cosine]
    L --> N[Index Type: HNSW]
    L --> O[Parameters: m=16, ef=100]
    
    A --> P[Search Operations]
    P --> Q[Vector Search]
    P --> R[Filtered Search]
    P --> S[Hybrid Search]
    
    style A fill:#e3f2fd
    style B fill:#f3e5f5
    style L fill:#fff3e0
    style P fill:#e8f5e8
```

## 8. 总结

本文档详细展示了AI记忆银行系统的完整架构设计和核心工作流程。系统采用了现代化的微服务架构设计原则，实现了以下关键特性：

### 8.1 架构优势

1. **分层清晰**: 严格的Clean Architecture实现确保了各层职责明确
2. **可扩展性**: 接口驱动设计支持功能模块的独立扩展  
3. **高性能**: 异步处理和智能缓存策略优化系统性能
4. **可观测性**: 完整的监控、日志和告警体系
5. **容错性**: 多层错误处理和恢复机制

### 8.2 技术创新点

1. **智能嵌入管理**: 自动化的向量嵌入生成和缓存
2. **混合搜索算法**: 文本搜索与语义搜索的智能结合
3. **双数据库支持**: PostgreSQL + Qdrant的灵活切换
4. **异步作业系统**: Redis队列驱动的高并发处理

### 8.3 生产就绪特性

1. **配置管理**: 多源配置和环境变量支持
2. **容器化**: Docker多阶段构建和部署
3. **监控告警**: Prometheus + Grafana监控栈
4. **日志聚合**: 结构化日志和ELK栈集成
5. **安全认证**: API Key认证和权限控制

该系统已具备了企业级AI记忆服务的完整能力，可直接用于生产环境部署和大规模应用。

---

**文档维护者**: AI Architecture Team  
**最后更新**: 2025-08-26  
**相关文档**: [当前实现分析](./current-implementation-analysis.md)